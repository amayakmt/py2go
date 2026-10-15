import ipaddress

a = ipaddress.ip_address("192.168.1.10")
print(a, repr(a), a.version, a.is_private, a.is_loopback, a.is_global, int(a), a + 5, a - 10, a.packed, a.reverse_pointer)
b = ipaddress.ip_address("2001:db8::1")
print(b, repr(b), b.version, b.exploded, b.compressed, b.is_private, int(b) >> 100, b.packed[:4], b.ipv4_mapped)
n = ipaddress.ip_network("10.0.0.0/22")
print(n, n.netmask, n.hostmask, n.broadcast_address, n.num_addresses, n.prefixlen, list(n.subnets(new_prefix=24)), n.supernet())
print(ipaddress.ip_address("10.0.1.7") in n, ipaddress.ip_address("10.0.4.7") in n, list(n.hosts())[:2], n[5], len(list(n)))
i = ipaddress.ip_interface("192.168.0.5/24")
print(i, i.ip, i.network, i.netmask, i.with_prefixlen, i.with_netmask)
print(ipaddress.ip_network("192.168.1.5/24", strict=False), sorted([ipaddress.ip_address("10.0.0.2"), ipaddress.ip_address("9.9.9.9")]))
print(list(ipaddress.summarize_address_range(ipaddress.ip_address("10.0.0.1"), ipaddress.ip_address("10.0.0.9"))))
print(list(ipaddress.collapse_addresses([ipaddress.ip_network("10.0.0.0/25"), ipaddress.ip_network("10.0.0.128/25")])))
print(ipaddress.ip_network("10.0.0.0/24").overlaps(ipaddress.ip_network("10.0.0.128/25")), ipaddress.ip_network("10.0.0.0/24").subnet_of(ipaddress.ip_network("10.0.0.0/16")))
print(ipaddress.IPv4Address("1.2.3.4") == ipaddress.ip_address("1.2.3.4"), hash(ipaddress.ip_address("1.2.3.4")) == hash(ipaddress.IPv4Address(0x01020304)))
for bad in ["256.1.1.1", "1.2.3", "::g", "1.2.3.4/33", "not an ip"]:
    for f in (ipaddress.ip_address, ipaddress.ip_network):
        try:
            f(bad)
        except ValueError as e:
            print(f.__name__, e)
try:
    ipaddress.IPv4Address("1.1.1.1") + (2**32)
except ipaddress.AddressValueError as e:
    print(e)
print(ipaddress.ip_network("::/0").num_addresses, ipaddress.IPv6Network("fe80::/64").is_link_local, ipaddress.ip_address("::ffff:1.2.3.4").ipv4_mapped)
