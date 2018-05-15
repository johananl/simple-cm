insert into simplecm.hosts (hostname, user) values ('172.28.128.3', 'vagrant');

insert into simplecm.operations (hostname, op_type, attributes)
    values ('172.28.128.3', 'FileExistsOperation',
    {'Description': 'verify_test_file_exists', 'Path': '/tmp/test.txt'});

insert into simplecm.operations (hostname, op_type, attributes)
    values ('172.28.128.3', 'FileContainsOperation',
    {'Description': 'verify_test_file_contains_hello', 'Path': '/tmp/test.txt', 'Text': 'hello'});