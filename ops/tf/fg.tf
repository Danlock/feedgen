provider "scaleway" {
  organization = "${var.org_id}"
  token        = "${var.api_token}"
  region       = "par1"
}

resource "scaleway_ip" "fg-ip" {}

resource "scaleway_server" "fg" {
  name           = "feedgen"
  image          = "${var.image_id}"
  type           = "${var.server_type}"
  security_group = "${scaleway_security_group.http.id}"
  cloudinit      = "${file("init.yml")}"
  public_ip      = "${scaleway_ip.fg-ip.ip}"
}

resource "scaleway_security_group" "http" {
  name        = "http"
  description = "allow HTTP and HTTPS traffic"
}

resource "scaleway_security_group_rule" "http_accept" {
  security_group = "${scaleway_security_group.http.id}"

  action    = "accept"
  direction = "inbound"
  ip_range  = "0.0.0.0/0"
  protocol  = "TCP"
  port      = 80
}

resource "scaleway_security_group_rule" "https_accept" {
  security_group = "${scaleway_security_group.http.id}"

  action    = "accept"
  direction = "inbound"
  ip_range  = "0.0.0.0/0"
  protocol  = "TCP"
  port      = 443
}
