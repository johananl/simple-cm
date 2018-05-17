create keyspace if not exists simplecm with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
create table if not exists simplecm.hosts(hostname text, user text, key_name text, password text, PRIMARY KEY(hostname));
create table if not exists simplecm.operations(hostname text, description text, script_name text, attributes map<text, text>, PRIMARY KEY(hostname, script_name));
-- TODO Set primary keys properly

insert into simplecm.hosts (hostname, user, key_name, password) values ('host1', 'root', '', 'root');
insert into simplecm.operations (hostname, description, script_name, attributes)
    values ('host1', 'verify_test_file_exists', 'file_exists', {'path': '/tmp/test.txt'});
insert into simplecm.operations (hostname, description, script_name, attributes)
    values ('host1', 'verify_test_file_contains_hello', 'file_contains', {'path': '/tmp/test.txt', 'text': 'hello'});