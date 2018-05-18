create keyspace if not exists simplecm with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
create table if not exists simplecm.hosts(hostname text, user text, key_name text, password text, PRIMARY KEY(hostname));
create table if not exists simplecm.operations_by_hostname(hostname text, description text, script_name text, attributes map<text, text>, PRIMARY KEY(hostname, script_name));

insert into simplecm.hosts (hostname, user, key_name, password) values ('host1', 'root', '', 'root');
insert into simplecm.hosts (hostname, user, key_name, password) values ('host2', 'root', '', 'root');
insert into simplecm.hosts (hostname, user, key_name, password) values ('host3', 'root', '', 'root');
insert into simplecm.hosts (hostname, user, key_name, password) values ('host4', 'root', '', 'root');
insert into simplecm.hosts (hostname, user, key_name, password) values ('host5', 'root', '', 'root');

insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host1', 'verify_test_file_exists', 'file_exists', {'path': '/etc/hosts'});
insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host1', 'verify_test_file_contains_127.0.0.1', 'file_contains', {'path': '/etc/hosts', 'text': '127.0.0.1'});

insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host2', 'verify_test_file_exists', 'file_exists', {'path': '/etc/hosts'});
insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host2', 'verify_test_file_contains_127.0.0.1', 'file_contains', {'path': '/etc/hosts', 'text': '127.0.0.1'});

insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host3', 'verify_test_file_exists', 'file_exists', {'path': '/etc/hosts'});
insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host3', 'verify_test_file_contains_127.0.0.1', 'file_contains', {'path': '/etc/hosts', 'text': '127.0.0.1'});

insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host4', 'verify_test_file_exists', 'file_exists', {'path': '/etc/hosts'});
insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host4', 'verify_test_file_contains_127.0.0.1', 'file_contains', {'path': '/etc/hosts', 'text': '127.0.0.1'});

insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host5', 'verify_test_file_exists', 'file_exists', {'path': '/etc/hosts'});
insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host5', 'verify_test_file_contains_127.0.0.1', 'file_contains', {'path': '/etc/hosts', 'text': '127.0.0.1'});
