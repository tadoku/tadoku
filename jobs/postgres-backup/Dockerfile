FROM library/postgres:14.5
RUN apt update -y
RUN apt install -y awscli curl
COPY backup.sh /
RUN chmod +x /backup.sh
