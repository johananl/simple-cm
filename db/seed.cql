-- Start from scratch to make this script idempotent.
drop keyspace if exists simplecm;

-- Create the keyspace.
create keyspace if not exists simplecm with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

-- Satisfies query: "get a host by hostname". Hostnames are unique.
create table if not exists simplecm.hosts(hostname text, user text, key_name text, password text, primary key(hostname));

-- Satisfies query: "get all operations for a hostname". An ID is added for row uniqueness since we could have more than one operation for the same hostname.
create table if not exists simplecm.operations(id UUID, hostname text, description text, script_name text, attributes map<text, text>, primary key(hostname, id));

-- Satisfies query: "get a run by its ID". Create time is defined as a clustering key to allow easy retrievals of runs for a given time frame.
create table if not exists simplecm.runs(id UUID, create_time timestamp, primary key(id, create_time));

-- Satisfies query: "get all results for a run".
create table if not exists simplecm.results_by_run_id(id UUID, run_id UUID, hostname text, ts timestamp, script_name text, successful boolean, primary key(run_id, id));
-- Satisfies query: "get all results for a run and a hostname".
-- TODO Do we need both results tables?
create table if not exists simplecm.results_by_run_id_and_hostname(id UUID, run_id UUID, hostname text, ts timestamp, script_name text, successful boolean, primary key(run_id, hostname, id));

-- Insert dummy data.
insert into simplecm.hosts (hostname, user, key_name, password) values ('host1', 'root', '', 'root');
insert into simplecm.hosts (hostname, user, key_name, password) values ('host2', 'root', '', 'root');
insert into simplecm.hosts (hostname, user, key_name, password) values ('host3', 'root', '', 'root');
insert into simplecm.hosts (hostname, user, key_name, password) values ('host4', 'root', '', 'root');
insert into simplecm.hosts (hostname, user, key_name, password) values ('host5', 'root', '', 'root');

insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host1', 'verify_test_file_exists', 'file_exists', {'path': '/etc/passwd'});
insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host1', 'verify_test_file_contains_1.1.1.1', 'file_contains', {'path': '/etc/hosts', 'text': '1.1.1.1 cloudflare-dns'});
insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host1', 'sleep', 'sleep', {'seconds': '10'});

insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host2', 'verify_test_file_exists', 'file_exists', {'path': '/etc/passwd'});
insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host2', 'verify_test_file_contains_1.1.1.1', 'file_contains', {'path': '/etc/hosts', 'text': '1.1.1.1 cloudflare-dns'});
insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host2', 'sleep', 'sleep', {'seconds': '10'});

insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host3', 'verify_test_file_exists', 'file_exists', {'path': '/etc/passwd'});
insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host3', 'verify_test_file_contains_1.1.1.1', 'file_contains', {'path': '/etc/hosts', 'text': '1.1.1.1 cloudflare-dns'});
insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host3', 'sleep', 'sleep', {'seconds': '10'});

insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host4', 'verify_test_file_exists', 'file_exists', {'path': '/etc/passwd'});
insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host4', 'verify_test_file_contains_1.1.1.1', 'file_contains', {'path': '/etc/hosts', 'text': '1.1.1.1 cloudflare-dns'});
insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host4', 'sleep', 'sleep', {'seconds': '10'});

insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host5', 'verify_test_file_exists', 'file_exists', {'path': '/etc/passwd'});
insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host5', 'verify_test_file_exists', 'file_exists', {'path': '/etc/inittab'});
insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host5', 'verify_test_file_contains_1.1.1.1', 'file_contains', {'path': '/etc/hosts', 'text': '1.1.1.1 cloudflare-dns'});
insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host5', 'this_operation_should_fail', 'file_contains', {'path': '/etc/wrong', 'text': 'oops'});
insert into simplecm.operations (id, hostname, description, script_name, attributes) values (uuid(), 'host5', 'sleep', 'sleep', {'seconds': '10'});