create keyspace if not exists simplecm with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
create table if not exists simplecm.hosts(id UUID, hostname text, user text, PRIMARY KEY(id));
create index if not exists on simplecm.hosts(hostname);