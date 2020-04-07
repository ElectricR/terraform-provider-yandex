---
layout: "yandex"
page_title: "Yandex: yandex_vpc_security_group"
sidebar_current: "docs-yandex-vpc-security-group"
description: |-
  Yandex VPC security group.
---

# yandex\_vpc\_security\_group

Manages a security group within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/vpc/concepts).

Security groups is in private preview phase and not available right now. Please wait for public preview announcement.


## Example Usage

```hcl
resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}

resource "yandex_vpc_security_group" "group1" {
  name        = "My security group"
  description = "description for my security group"
  network_id  = "${yandex_vpc_network.lab-net.id}"

  labels = {
    my-label    = "my-label-value"
  }

  rule {
    direction = "INGRESS"
    description = "rule1 description"
	v4_cidr_blocks = ["10.0.0.1/24", "10.0.0.2/24"]
    port = 8080
  }

  rule {
    direction = "EGRESS"
    description = "rule2 description"
	v4_cidr_blocks = ["10.0.0.1/24", "10.0.0.2/24"]
    from_port = 8090
    to_port = 8099
  }
}
```

## Argument Reference

The following arguments are supported:

* `network_id` (Required) - ID of the network this security group belongs to.

---

* `name` (Optional) - Name of the security group.
* `description` (Optional) - Description of the security group.
* `folder_id` (Optional) - ID of the folder this security group belongs to.
* `labels` (Optional) - Labels to assign to this security group.
* `rule` (Optional) - A list of rules.

The structure is documented below.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - ID of the route table to assign to this security group.
* `created_at` - Creation timestamp of this security group.

---

The `rule` block supports:

* `direction` (Required) - Direction of the rule. Values are `INGRESS` or `EGRESS`

---

* `description` (Optional) - Description of the rule.
* `labels` (Optional) - Labels to assign to this rule.
* `protocol_name` (Optional) -  Name of the protocol. Values are `tcp`, `udp`, `icmp`, `any`.
* `protocol_number` (Optional) - Number of the protocol defined by [IANA](https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml). Values are `0`,`6`,`17`.
* `from_port` (Optional) - Minimum port number. Values
* `to_port` (Optional) - Maximum port number.
* `port` (Optional) - Port number (if applied to a single port).
* `v4_cidr_blocks` (Optional) - The blocks of IPv4 addresses for this rule.
* `v6_cidr_blocks` (Optional) - The blocks of IPv6 addresses for this rule. `v6_cidr_blocks` argument is currently not supported. It will be available in the future.


~> **NOTE:** Only one of `protocol_name` or `protocol_number` can be specified. If non of them is set all protocols are allowed. Either `port` argument or `from_port` and `to_port` arguments together can be specified.


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - Id of the rule.