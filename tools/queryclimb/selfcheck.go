package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
)

func selfcheck(ctx context.Context, root string, args []string) error {
	flags := flag.NewFlagSet("selfcheck", flag.ContinueOnError)
	database := flags.String("name", "qc_schema", "migrated fixture database")
	if err := flags.Parse(args); err != nil {
		return err
	}
	conn, err := connect(ctx, localDSN, *database, "tadoku")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	var bypass, super, owns bool
	err = conn.QueryRow(ctx, `select r.rolbypassrls, r.rolsuper,
  exists(select 1 from pg_class where relnamespace = 'public'::regnamespace and relowner = r.oid)
  from pg_roles r where rolname = current_user`).Scan(&bypass, &super, &owns)
	if err != nil {
		return err
	}
	if bypass || super || owns {
		return errors.New("application role has owner privileges")
	}
	err = scoped(ctx, conn, "climb/other", func() error {
		var leaked int
		if err := conn.QueryRow(ctx, "select count(*) from users where tenant <> 'climb/other'").Scan(&leaked); err != nil {
			return err
		}
		if leaked != 0 {
			return errors.New("tenant leakage")
		}
		return nil
	})
	if err != nil {
		return err
	}

	a, b, other := "a", "b", "other"
	one := [][]*string{{&a, &b}, {&a, nil}, {&other, &b}}
	reorderedTie := [][]*string{{&a, nil}, {&a, &b}, {&other, &b}}
	wrongOrder := [][]*string{{&other, &b}, {&a, nil}, {&a, &b}}
	first, err := fingerprint(one, []int{0}, nil)
	if err != nil {
		return err
	}
	tied, _ := fingerprint(reorderedTie, []int{0}, nil)
	wrong, _ := fingerprint(wrongOrder, []int{0}, nil)
	duplicate, _ := fingerprint(append(one, one[0]), []int{0}, nil)
	if first != tied || first == wrong || first == duplicate {
		return errors.New("fingerprint ordering/multiplicity check failed")
	}

	q := query{
		Key: "selfcheck.Write",
		SQL: `update users
   set display_name = display_name || ' changed'
   where id = $1
   returning display_name`,
		Write: true, Tables: []string{"users"}, Mutations: []string{"users"},
	}
	c := caseSpec{
		Name: "rollback", Query: q.Key, Tenant: "tadoku/prod",
		Args: []*string{new(string)},
	}
	if err := conn.QueryRow(ctx, "select md5('user-1')::uuid::text").Scan(c.Args[0]); err != nil {
		return err
	}
	result, err := measure(ctx, conn, q, c)
	if err != nil {
		return err
	}
	if result.Affected != 1 || result.CustomPlans+result.GenericPlans != 6 {
		return errors.New("write execution check failed")
	}
	err = scoped(ctx, conn, "tadoku/prod", func() error {
		var name string
		err := conn.QueryRow(ctx, "select display_name from users where id = $1::uuid", *c.Args[0]).Scan(&name)
		if err != nil {
			return err
		}
		if name != "Learner 1" {
			return fmt.Errorf("write survived rollback: %s", name)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return output(map[string]any{"role": "tadoku", "tenant_isolation": true, "write_rollback": true,
		"sixth_execution": true, "fingerprint_ordering": true})
}
