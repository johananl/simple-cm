#!/bin/bash

for i in $(terraform output -state ../testing/terraform/terraform.tfstate -json | jq -r '.public_ips.value[]'); do
    echo "insert into simplecm.hosts (hostname, user) values ('$i', 'ec2-user');"
done

for i in $(terraform output -state ../testing/terraform/terraform.tfstate -json | jq -r '.public_ips.value[]'); do
echo "insert into simplecm.operations (hostname, description, script_name, attributes)
    values ('$i', 'verify_test_file_exists', 'file_exists', {'path': '/tmp/test.txt'});
insert into simplecm.operations (hostname, description, script_name, attributes)
    values ('$i', 'verify_test_file_contains_hello', 'file_contains', {'path': '/tmp/test.txt', 'text': 'hello'});";
done
