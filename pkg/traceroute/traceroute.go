package traceroute

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"
	"time"

	"golang.org/x/net/ipv4"
)

const DEFAULT_PORT = 33434
const DEFAULT_MAX_HOPS = 64
const DEFAULT_FIRST_HOP = 1
const DEFAULT_TIMEOUT_MS = 3000
const DEFAULT_RETRIES = 3
const DEFAULT_PACKET_SIZE = 52

type Hop struct {
	IP      netip.Addr
	Latency time.Duration
	TTL     uint16
}

type Tracer struct {
	Destiny  netip.Addr
	Hops     []*Hop
	RecvConn net.PacketConn

	ctx    context.Context
	cancel context.CancelFunc
	timeCh chan time.Time
	ttlCh  chan int
	errCh  chan error
}

func (t *Tracer) Start() error {

	go func() {
		defer t.RecvConn.Close()
		for t.ctx.Err() == nil {
			err := t.RecvConn.SetReadDeadline(time.Now().Add(time.Second * 3))
			if err != nil {
				t.errCh <- err
				return
			}
			ttl := len(t.Hops) + 1
			if ttl >= DEFAULT_MAX_HOPS {
				t.cancel()
				return
			}
			t.ttlCh <- len(t.Hops) + 1
			select {
			case <-t.ctx.Done():
				return
			case begin := <-t.timeCh:
				buf := make([]byte, DEFAULT_PACKET_SIZE)
				addr := "*"
				_, _addr, err := t.RecvConn.ReadFrom(buf)
				if err != nil {
					if !os.IsTimeout(err) {
						t.errCh <- err
						return
					}
				} else {
					addr = _addr.String()
				}

				end := time.Since(begin)
				fmt.Printf("[%v] %s\n", end, addr)
				t.Hops = append(t.Hops, &Hop{})
				if addr == t.Destiny.String() {
					t.cancel()
				}
			}
		}
	}()

	go func() {
		for t.ctx.Err() == nil {
			ttl := -1
			select {
			case <-t.ctx.Done():
				return
			case t := <-t.ttlCh:
				ttl = t
			}

			udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
			if err != nil {
				t.errCh <- err
				return
			}
			defer udpConn.Close()

			pktConn := ipv4.NewPacketConn(udpConn)
			pktConn.SetTTL(ttl)

			buf := make([]byte, DEFAULT_PACKET_SIZE)
			addrPort := netip.AddrPortFrom(t.Destiny, DEFAULT_PORT)
			_, err = udpConn.WriteToUDPAddrPort(buf, addrPort)
			if err != nil {
				t.errCh <- err
				return
			}
			t.timeCh <- time.Now()
		}
	}()

	select {
	case <-t.ctx.Done():
		return nil
	case err := <-t.errCh:
		t.cancel()
		return err
	}

}

func (t *Tracer) Close() error {
	if err := t.RecvConn.Close(); err != nil {
		return err
	}
	if t.ctx.Err() == nil {
		t.cancel()
	}
	return nil
}

func NewTracer(_ctx context.Context, address string) (*Tracer, error) {
	ctx, cancel := context.WithCancel(_ctx)

	tdest := net.ParseIP(address)
	if tdest == nil {
		ips, err := net.LookupIP(address)
		if err != nil {
			cancel()
			return nil, err
		}
		for _, v := range ips {
			if v.To4() == nil {
				continue
			}
			tdest = v
			break
		}
	}

	dest, err := netip.ParseAddr(tdest.String())
	if err != nil {
		cancel()
		return nil, err
	}

	listenCfg := net.ListenConfig{
		Control:   nil,
		KeepAlive: -1,
	}

	icmpConn, err := listenCfg.ListenPacket(ctx, "ip4:icmp", "0.0.0.0")
	if err != nil {
		cancel()
		return nil, err
	}

	tracer := Tracer{
		Destiny:  dest,
		Hops:     []*Hop{},
		RecvConn: icmpConn,
		ctx:      ctx,
		cancel:   cancel,
		timeCh:   make(chan time.Time),
		ttlCh:    make(chan int),
		errCh:    make(chan error),
	}

	return &tracer, nil
}
