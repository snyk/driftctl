# data source for an official Ubuntu 20.04 AMI
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

# # a simple aws instance
resource "aws_instance" "test_instance_1" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"
  #   key_name = aws_key_pair.keypair1.key_name

  tags = {
    Name = "test_instance_1"
  }

  root_block_device {
    volume_type           = "gp2"
    volume_size           = 20
    delete_on_termination = true
  }
}
