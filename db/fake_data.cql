insert into simplecm.hosts (hostname, user) values ('172.28.128.3', 'vagrant');

insert into simplecm.operations (hostname, description, script_name, attributes)
    values ('172.28.128.3', 'verify_test_file_exists', 'file_exists', {'path': '/tmp/test.txt'});

insert into simplecm.operations (hostname, description, script_name, attributes)
    values ('172.28.128.3', 'verify_test_file_contains_hello', 'file_contains', {'path': '/tmp/test.txt', 'text': 'hello'});