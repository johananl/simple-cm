output "public_ips" {
  value = "${aws_instance.instances.*.public_ip}"
}
