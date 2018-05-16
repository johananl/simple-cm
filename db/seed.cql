create keyspace if not exists simplecm with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
create table if not exists simplecm.hosts(hostname text, user text, PRIMARY KEY(hostname));
create table if not exists simplecm.operations(hostname text, description text, script_name text, attributes map<text, text>, PRIMARY KEY(hostname, script_name));
-- TODO Set primary keys properly