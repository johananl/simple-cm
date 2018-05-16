provider "aws" {
  region  = "${var.region}"
  profile = "${var.profile}"
}

data "aws_ami" "amazon_linux" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-*-x86_64-gp2"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }
}

resource "aws_security_group" "instances" {
  name = "CM-Test"
}

resource "aws_security_group_rule" "ssh" {
  security_group_id = "${aws_security_group.instances.id}"
  type              = "ingress"
  protocol          = "tcp"
  from_port         = 22
  to_port           = 22
  cidr_blocks       = ["${var.ssh_source_range}"]
}

resource "aws_security_group_rule" "outbound" {
  security_group_id = "${aws_security_group.instances.id}"
  type              = "egress"
  protocol          = -1
  from_port         = 0
  to_port           = 65535
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_instance" "instances" {
  count                       = "${var.num_of_instances}"
  ami                         = "${data.aws_ami.amazon_linux.id}"
  instance_type               = "${var.instance_type}"
  key_name                    = "${var.key_name}"
  associate_public_ip_address = "${var.public_ip}"
  vpc_security_group_ids      = ["${aws_security_group.instances.id}"]

  tags {
    Name = "CM-Test-${count.index}"
  }
}
