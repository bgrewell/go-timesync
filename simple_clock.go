package timesync

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

type SimpleClock struct {
	adjustment       time.Duration
	delay            time.Duration
	offset           time.Duration
	interval         time.Duration
	adjustSystemTime bool
	syncTime         bool
	serveTime        bool
	lastSync         time.Time
	host             string
	port             int
	pool             []int64
	poolLen          int
	idx              int64
}

func (sc *SimpleClock) sync() error {
	// Send a single sync first to make sure we don't have a problem
	if err := sc.sendSync(true); err != nil {
		return err
	}
	sc.sendSync(true)

	// Start period syncs
	go sc.periodicSync()
	return nil
}

func (sc *SimpleClock) periodicSync() {
	for sc.syncTime {
		err := sc.sendSync(false)
		if err != nil {
			fmt.Println("Error syncing:", err)
		}
		time.Sleep(sc.interval)
	}
}

func (sc *SimpleClock) sendSync(compensateInternalClock bool) error {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", sc.host, sc.port))
	if err != nil {
		return err
	}
	defer conn.Close()

	pkt := Payload{}
	pkt.T1 = sc.Now().UnixNano()
	err = binary.Write(conn, binary.BigEndian, pkt)
	if err != nil {
		return err
	}

	err = binary.Read(conn, binary.BigEndian, &pkt)
	if err != nil {
		return err
	}
	pkt.T4 = sc.Now().UnixNano()

	// Update offset and delay
	sc.offset = time.Duration(CalcOffset(pkt.T1, pkt.T2, pkt.T3, pkt.T4))
	sc.delay = time.Duration(CalcDelay(pkt.T1, pkt.T2, pkt.T3, pkt.T4))

	if compensateInternalClock {
		// Slowly adjust clock
		sc.calcAdjustment()
	}

	return nil
}

func (sc *SimpleClock) serve() error {
	addr := net.UDPAddr{
		IP:   net.ParseIP(sc.host),
		Port: sc.port,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	for sc.serveTime {

		buf := make([]byte, 32)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		T1 := binary.BigEndian.Uint64(buf[0:n])

		pkt := Payload{
			T1: int64(T1),
		}

		pkt.T2 = time.Now().UnixNano()
		txbuf := make([]byte, 32)
		binary.BigEndian.PutUint64(txbuf[0:8], uint64(pkt.T1))
		binary.BigEndian.PutUint64(txbuf[8:16], uint64(pkt.T2))
		pkt.T3 = time.Now().UnixNano()
		binary.BigEndian.PutUint64(txbuf[16:24], uint64(pkt.T3))

		_, err = conn.WriteToUDP(txbuf, addr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sc *SimpleClock) calcAdjustment() {
	sc.adjustment += sc.offset
}

func (sc SimpleClock) Now() time.Time {
	return time.Now().Add(sc.adjustment)
}

func (sc *SimpleClock) EnableSync(host string, port int) error {

	if sc.serveTime {
		return errors.New("you can not sync time while serving time to clients")
	}

	sc.syncTime = true
	sc.host = host
	sc.port = port
	sc.poolLen = 10
	sc.pool = make([]int64, sc.poolLen)
	return sc.sync()
}

func (sc *SimpleClock) DisableSync() {
	sc.syncTime = false
}

func (sc *SimpleClock) EnableSystemTimeAdjustment() {
	sc.adjustSystemTime = true
}

func (sc *SimpleClock) DisableSystemTimeAdjustment() {
	sc.adjustSystemTime = false
}

func (sc *SimpleClock) SetInterval(interval time.Duration) {
	sc.interval = interval
}

func (sc SimpleClock) GetInterval() time.Duration {
	return sc.interval
}

func (sc SimpleClock) LastSync() time.Time {
	return time.Now()
}

func (sc SimpleClock) Offset() time.Duration {
	return sc.offset
}

func (sc SimpleClock) Delay() time.Duration {
	return sc.delay
}

func (sc *SimpleClock) EnableServer(listenAddress string, port int) error {

	if sc.syncTime {
		return errors.New("you can not serve time to clients while synchronizing time with a server")
	}

	sc.serveTime = true
	sc.host = listenAddress
	sc.port = port
	return sc.serve()
}

func (sc *SimpleClock) DisableServer() {
	sc.serveTime = false
}
