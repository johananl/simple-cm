create keyspace if not exists simplecm with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
create table if not exists simplecm.hosts(hostname text, user text, key_name text, password text, PRIMARY KEY(hostname));
create table if not exists simplecm.operations_by_hostname(hostname text, description text, script_name text, attributes map<text, text>, PRIMARY KEY(hostname, script_name));

-- Satisfies query: "get a run by its ID". Create time is defined as a clustering key to allow easy retrievals of runs for a given time frame.
create table if not exists simplecm.runs(id UUID, create_time timestamp, primary key(id, create_time));

-- Satisfies query: "get all results for a run".
create table if not exists simplecm.results_by_run_id(id UUID, run_id UUID, hostname text, ts timestamp, script_name text, successful boolean, primary key(run_id, id));
-- Satisfies query: "get all results for a run and a hostname".
create table if not exists simplecm.results_by_run_id_and_hostname(id UUID, run_id UUID, hostname text, ts timestamp, script_name text, successful boolean, primary key(run_id, hostname, id));

insert into simplecm.hosts (hostname, user, key_name, password) values ('host1', 'root', '', 'root');
insert into simplecm.hosts (hostname, user, key_name, password) values ('host2', 'root', '', 'root');
insert into simplecm.hosts (hostname, user, key_name, password) values ('host3', 'root', '', 'root');
insert into simplecm.hosts (hostname, user, key_name, password) values ('host4', 'root', '', 'root');
insert into simplecm.hosts (hostname, user, key_name, password) values ('host5', 'root', '', 'root');

insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host1', 'verify_test_file_exists', 'file_exists', {'path': '/etc/hosts'});
insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host1', 'verify_test_file_contains_1.1.1.1', 'file_contains', {'path': '/etc/hosts', 'text': '1.1.1.1 cloudflare-dns'});

insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host2', 'verify_test_file_exists', 'file_exists', {'path': '/etc/hosts'});
insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host2', 'verify_test_file_contains_1.1.1.1', 'file_contains', {'path': '/etc/hosts', 'text': '1.1.1.1 cloudflare-dns'});

insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host3', 'verify_test_file_exists', 'file_exists', {'path': '/etc/hosts'});
insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host3', 'verify_test_file_contains_1.1.1.1', 'file_contains', {'path': '/etc/hosts', 'text': '1.1.1.1 cloudflare-dns'});

insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host4', 'verify_test_file_exists', 'file_exists', {'path': '/etc/hosts'});
insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host4', 'verify_test_file_contains_1.1.1.1', 'file_contains', {'path': '/etc/hosts', 'text': '1.1.1.1 cloudflare-dns'});

insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host5', 'verify_test_file_exists', 'file_exists', {'path': '/etc/hosts'});
insert into simplecm.operations_by_hostname (hostname, description, script_name, attributes) values ('host5', 'verify_test_file_contains_1.1.1.1', 'file_contains', {'path': '/etc/hosts', 'text': '1.1.1.1 cloudflare-dns'});
