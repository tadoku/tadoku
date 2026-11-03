package identity

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"net"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/uuidx"
	valkeygo "github.com/valkey-io/valkey-go"
)

func TestRegisterUsesPinnedUUID(t *testing.T) {
	t.Parallel()

	wantBytes := bytes.Repeat([]byte{0x11}, 16)
	want := uuid.Must(uuid.NewRandomFromReader(bytes.NewReader(wantBytes)))
	ctx := uuidx.WithSource(
		tenant.WithKey(t.Context(), tenant.Production()),
		bytes.NewReader(wantBytes),
	)
	service := NewService(
		NewRepository(admittingPostgres(t)),
		NewThrottle(admittingValkey(t), time.Second),
		"tadoku.app",
		nil,
	)

	result, err := service.Register(ctx, Registration{
		Email:       "reader@example.test",
		DisplayName: "Reader",
		Password:    "Independent-653!",
		IP:          "192.0.2.1",
	})

	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if result.Session.Account.ID != want {
		t.Errorf("account ID = %s, want pinned %s", result.Session.Account.ID, want)
	}
}

type admittingValkeyClient struct{ valkeygo.Client }

func admittingValkey(t *testing.T) valkeygo.Client {
	t.Helper()

	base, err := valkeygo.NewClient(valkeygo.ClientOption{
		InitAddress:       []string{"127.0.0.1:1"},
		ForceSingleClient: true,
		DisableCache:      true,
		Dialer:            net.Dialer{Timeout: time.Millisecond},
		DialCtxFn: func(context.Context, string, *net.Dialer, *tls.Config) (net.Conn, error) {
			return nil, errors.New("admitting valkey does not dial")
		},
	})
	if base == nil {
		t.Fatalf("ForceSingleClient valkey: %v", err)
	}
	t.Cleanup(base.Close)
	return admittingValkeyClient{Client: base}
}

func (admittingValkeyClient) Do(context.Context, valkeygo.Completed) valkeygo.ValkeyResult {
	buf := make([]byte, 7)
	buf = append(buf, '*')
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], 1)
	buf = append(buf, length[:]...)
	buf = append(buf, ':')
	binary.BigEndian.PutUint64(length[:], 1)
	buf = append(buf, length[:]...)

	var message valkeygo.ValkeyMessage
	if err := message.CacheUnmarshalView(buf); err != nil {
		return valkeygo.NewErrorResult(err)
	}
	return valkeygo.NewResult(message, nil)
}

var sqlPlaceholder = regexp.MustCompile(`\$(\d+)`)

func admittingPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}
			go admitPostgresConn(connection)
		}
	}()

	pool, err := pgxpool.New(
		t.Context(),
		"postgres://postgres:postgres@"+listener.Addr().String()+"/postgres?sslmode=disable",
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func admitPostgresConn(connection net.Conn) {
	defer connection.Close()
	_ = connection.SetDeadline(time.Now().Add(30 * time.Second))

	backend := pgproto3.NewBackend(connection, connection)
	if _, err := backend.ReceiveStartupMessage(); err != nil {
		return
	}
	backend.Send(&pgproto3.AuthenticationOk{})
	backend.Send(&pgproto3.BackendKeyData{ProcessID: 1, SecretKey: []byte{0, 0, 0, 1}})
	backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
	if err := backend.Flush(); err != nil {
		return
	}

	parameters := map[string][]uint32{}
	for {
		message, err := backend.Receive()
		if err != nil {
			return
		}
		switch typed := message.(type) {
		case *pgproto3.Parse:
			oids := typed.ParameterOIDs
			if len(oids) == 0 {
				max := 0
				for _, match := range sqlPlaceholder.FindAllStringSubmatch(typed.Query, -1) {
					index, err := strconv.Atoi(match[1])
					if err == nil && index > max {
						max = index
					}
				}
				oids = make([]uint32, max)
			}
			parameters[typed.Name] = oids
			backend.Send(&pgproto3.ParseComplete{})
		case *pgproto3.Bind:
			backend.Send(&pgproto3.BindComplete{})
		case *pgproto3.Describe:
			if typed.ObjectType == 'S' {
				backend.Send(&pgproto3.ParameterDescription{ParameterOIDs: parameters[typed.Name]})
			}
			backend.Send(&pgproto3.NoData{})
		case *pgproto3.Execute:
			backend.Send(&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")})
		case *pgproto3.Sync:
			backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
			if err := backend.Flush(); err != nil {
				return
			}
		case *pgproto3.Query:
			backend.Send(&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")})
			backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
			if err := backend.Flush(); err != nil {
				return
			}
		case *pgproto3.Close:
			backend.Send(&pgproto3.CloseComplete{})
		case *pgproto3.Flush:
			if err := backend.Flush(); err != nil {
				return
			}
		case *pgproto3.Terminate:
			return
		}
	}
}
