//go:build !linux

package udp

import "gnalloy.org/gnalloy/transport"

type datagramBatchWriter struct{}

func sendDatagramBatch(fd transport.FDRef, datagrams []Datagram) (int, bool, error) {
	return new(datagramBatchWriter).send(fd, datagrams)
}

func (*datagramBatchWriter) send(fd transport.FDRef, datagrams []Datagram) (int, bool, error) {
	for index := range datagrams {
		again, err := sendDatagram(fd, datagrams[index])
		if err != nil {
			return index, false, err
		}
		if again {
			return index, true, nil
		}
	}
	return len(datagrams), false, nil
}
