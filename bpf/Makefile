
CC=clang
CFLAGS=-O2 -g  -Wall -target bpf -I/usr/include/$(shell uname -m)-linux-gnu

BPF_FS=/sys/fs/bpf
PROG_MOUNT_PATH=/sys/fs/bpf

MACROS:=
DEBUG ?= 1
BRIDGE_ETH ?= cni0
BRIDGE_IP ?= 183763201 #10.244.1.1

# see https://stackoverflow.com/questions/15063298/how-to-check-kernel-version-in-makefile
KVER = $(shell uname -r)
KMAJ = $(shell echo $(KVER) | \
sed -e 's/^\([0-9][0-9]*\)\.[0-9][0-9]*\.[0-9][0-9]*.*/\1/')
KMIN = $(shell echo $(KVER) | \
sed -e 's/^[0-9][0-9]*\.\([0-9][0-9]*\)\.[0-9][0-9]*.*/\1/')
KREV = $(shell echo $(KVER) | \
sed -e 's/^[0-9][0-9]*\.[0-9][0-9]*\.\([0-9][0-9]*\).*/\1/')

kver_ge = $(shell \
echo test | awk '{if($(KMAJ) < $(1)) {print 0} else { \
if($(KMAJ) > $(1)) {print 1} else { \
if($(KMIN) < $(2)) {print 0} else { \
if($(KMIN) > $(2)) {print 1} else { \
if($(KREV) < $(3)) {print 0} else { print 1 } \
}}}}}' \
)

# See https://nakryiko.com/posts/bpf-tips-printk/, kernel will auto print newline if version greater than 5.9.0
ifneq ($(call kver_ge,5,8,999),1)
MACROS:= $(MACROS) -DPRINTNL # kernel version less
endif

MACROS:= $(MACROS) -DBRIDGE_IP=$(BRIDGE_IP)

ifeq ($(DEBUG),1)
    MACROS:= $(MACROS) -DDEBUG
endif

CGROUP2_PATH ?= $(shell mount | grep cgroup2 | awk '{print $$3}' | grep -v "^/host" | head -n 1)
ifeq ($(CGROUP2_PATH),)
$(error It looks like your system does not have cgroupv2 enabled, or the automatic recognition fails. Please enable cgroupv2, or specify the path of cgroupv2 manually via CGROUP2_PATH parameter.)
endif

TARGETS=ecnet_dnsproxy.o ecnet_connect.o ecnet_redirmsg.o ecnet_sockops.o ecnet_sockopt.o
$(BPF_FS):
	mountpoint -q $@ || mount -t bpf bpf $@

init-bpffs:
	mount -t bpf bpf ${BPF_FS}
	mkdir -p $(PROG_MOUNT_PATH)/xdp/globals

%.o: %.c
	$(CC) $(CFLAGS) $(MACROS) -c $< -o $@

generate-compilation-database:
	CC="$(CC)" CFLAGS="$(CFLAGS)" MACROS="$(MACROS)" scripts/generate-compilation-database.sh | tee compile_commands.json

compile: $(TARGETS)

compile-clean:
	[ -f ecnet_connect.c ] && sudo rm -f $(TARGETS)

# Map
load-map-ecnet_dns_resdb:
	[ -f $(PROG_MOUNT_PATH)/ecnet_dns_resdb ] || sudo bpftool map create $(PROG_MOUNT_PATH)/ecnet_dns_resdb type lru_hash key 260 value 8 entries 65535 name ecnet_dns_resdb

load-map-ecnet_dns_endpt:
	[ -f $(PROG_MOUNT_PATH)/ecnet_dns_endpt ] || sudo bpftool map create $(PROG_MOUNT_PATH)/ecnet_dns_endpt type lru_hash key 4 value 4 entries 65535 name ecnet_dns_endpt

load-map-ecnet_sess_dest:
	[ -f $(PROG_MOUNT_PATH)/ecnet_sess_dest ] || sudo bpftool map create $(PROG_MOUNT_PATH)/ecnet_sess_dest type lru_hash key 8 value 8 entries 65535 name ecnet_sess_dest

load-map-ecnet_pair_dest:
	[ -f $(PROG_MOUNT_PATH)/ecnet_pair_dest ] || sudo bpftool map create $(PROG_MOUNT_PATH)/ecnet_pair_dest type lru_hash key 12 value 8 entries 65535 name ecnet_pair_dest

load-map-ecnet_sock_pair:
	[ -f $(PROG_MOUNT_PATH)/ecnet_sock_pair ] || sudo bpftool map create $(PROG_MOUNT_PATH)/ecnet_sock_pair type sockhash key 12 value 4 entries 65535 name ecnet_sock_pair

clean-maps:
	sudo rm -f \
		$(PROG_MOUNT_PATH)/ecnet_dns_resdb \
		$(PROG_MOUNT_PATH)/ecnet_dns_endpt \
		$(PROG_MOUNT_PATH)/ecnet_sess_dest \
		$(PROG_MOUNT_PATH)/ecnet_pair_dest \
		$(PROG_MOUNT_PATH)/ecnet_sock_pair \
		$(PROG_MOUNT_PATH)/sockops \
		$(PROG_MOUNT_PATH)/sockopt

load-dnsproxy: load-map-ecnet_dns_resdb load-map-ecnet_dns_endpt
	sudo bpftool -m prog loadall ecnet_dnsproxy.o $(PROG_MOUNT_PATH)/dnxproxy \
		map name ecnet_dns_resdb pinned $(PROG_MOUNT_PATH)/ecnet_dns_resdb \
		map name ecnet_dns_endpt pinned $(PROG_MOUNT_PATH)/ecnet_dns_endpt

attach-dnsproxy:
	sudo ip link set dev $(BRIDGE_ETH) xdpgeneric off
	sudo ip link set dev $(BRIDGE_ETH) xdpgeneric pinned $(PROG_MOUNT_PATH)/dnxproxy/xdp

clean-dnsproxy:
	sudo ip link set dev $(BRIDGE_ETH) xdpgeneric off
	sudo rm -rf $(PROG_MOUNT_PATH)/dnxproxy

load-connect: load-map-ecnet_sess_dest
	sudo bpftool -m prog loadall ecnet_connect.o $(PROG_MOUNT_PATH)/connect \
		map name ecnet_sess_dest pinned $(PROG_MOUNT_PATH)/ecnet_sess_dest

attach-connect:
	sudo bpftool cgroup attach $(CGROUP2_PATH) connect4 pinned $(PROG_MOUNT_PATH)/connect/cgroup_connect4

clean-connect:
	sudo bpftool cgroup detach $(CGROUP2_PATH) connect4 pinned $(PROG_MOUNT_PATH)/connect/cgroup_connect4
	sudo rm -rf $(PROG_MOUNT_PATH)/connect

load-redirmsg: load-map-ecnet_sock_pair
	sudo bpftool -m prog load ecnet_redirmsg.o $(PROG_MOUNT_PATH)/redirmsg \
		map name ecnet_sock_pair pinned $(PROG_MOUNT_PATH)/ecnet_sock_pair

attach-redirmsg:
	sudo bpftool prog attach pinned $(PROG_MOUNT_PATH)/redirmsg msg_verdict pinned $(PROG_MOUNT_PATH)/ecnet_sock_pair

clean-redirmsg:
	sudo bpftool prog detach pinned $(PROG_MOUNT_PATH)/redirmsg msg_verdict pinned $(PROG_MOUNT_PATH)/ecnet_sock_pair
	sudo rm $(PROG_MOUNT_PATH)/redirmsg

load-sockopt: load-map-ecnet_pair_dest
	sudo bpftool -m prog load ecnet_sockopt.o $(PROG_MOUNT_PATH)/sockopt \
		map name ecnet_pair_dest pinned $(PROG_MOUNT_PATH)/ecnet_pair_dest

attach-sockopt:
	sudo bpftool cgroup attach $(CGROUP2_PATH) getsockopt pinned $(PROG_MOUNT_PATH)/sockopt

clean-sockopt:
	sudo bpftool cgroup detach $(CGROUP2_PATH) getsockopt pinned $(PROG_MOUNT_PATH)/sockopt
	sudo rm $(PROG_MOUNT_PATH)/sockopt

load-sockops: load-map-ecnet_sess_dest load-map-ecnet_pair_dest load-map-ecnet_sock_pair
	sudo bpftool -m prog load ecnet_sockops.o $(PROG_MOUNT_PATH)/sockops \
	map name ecnet_sess_dest pinned $(PROG_MOUNT_PATH)/ecnet_sess_dest \
	map name ecnet_pair_dest pinned $(PROG_MOUNT_PATH)/ecnet_pair_dest \
	map name ecnet_sock_pair pinned $(PROG_MOUNT_PATH)/ecnet_sock_pair

attach-sockops:
	sudo bpftool cgroup attach $(CGROUP2_PATH) sock_ops pinned $(PROG_MOUNT_PATH)/sockops

clean-sockops:
	sudo bpftool cgroup detach $(CGROUP2_PATH) sock_ops pinned $(PROG_MOUNT_PATH)/sockops
	sudo rm -rf $(PROG_MOUNT_PATH)/sockops

load: compile init-bpffs load-dnsproxy load-connect load-redirmsg load-sockopt load-sockops

attach: attach-dnsproxy attach-connect attach-redirmsg attach-sockopt attach-sockops

clean: clean-dnsproxy clean-connect clean-redirmsg clean-sockopt clean-sockops clean-maps compile-clean

all: compile

clean-progs: compile-clean clean-maps
	sudo rm -rf $(PROG_MOUNT_PATH)/connect
	sudo rm -rf $(PROG_MOUNT_PATH)/redirmsg
	sudo rm -rf $(PROG_MOUNT_PATH)/sockops
	sudo rm -rf $(PROG_MOUNT_PATH)/sockopt