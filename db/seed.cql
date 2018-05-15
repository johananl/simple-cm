create keyspace if not exists simplecm with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
create table if not exists simplecm.hosts(hostname text, user text, PRIMARY KEY(hostname));
create table if not exists simplecm.operations(hostname text, op_type text, attributes map<text, text>, PRIMARY KEY(hostname, op_type));