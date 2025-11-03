package app

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"syscall"
)

type Reserved struct {
	Listener net.Listener
	Port     int
}

// ListenOnAvailablePort tạo net.Listener theo host/port.
//   - Nếu port == 0 ➜ OS tự chọn cổng trống.
//   - Nếu port > 0 mà bị "already in use" ➜ thử lại với port 0.
//
// Trả về listener & cổng thực sự đã bind.
func ListenOnAvailablePort(host string, port int) (*Reserved, error) {
	// 1️⃣  Nếu port==0 ‒ lấy cổng rảnh trước, rồi bind thật
	if port == 0 {
		r, err := FindAvailablePort(host)
		if err != nil {
			return nil, err
		}

		return r, nil
	}

	tryBind := func(p int) (net.Listener, error) {
		return net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(p)))
	}

	// 2️⃣  Thử bind cổng (đã xác định ở bước 1)
	l, err := tryBind(port)
	if err == nil {
		return &Reserved{Listener: l, Port: port}, nil
	}

	// 3️⃣  Nếu cổng bận / không đủ quyền ➜ fallback một lần
	if errors.Is(err, syscall.EADDRINUSE) || errors.Is(err, syscall.EACCES) {
		r, errFind := FindAvailablePort(host)
		if errFind != nil {
			return nil, errFind
		}
		return r, nil
	}

	return nil, fmt.Errorf("listen failed: %w", err)
}

func FindAvailablePort(host string) (*Reserved, error) {
	l, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		l.Close() // close immediately
		// Handle common “address not available” / permission errors explicitly
		if errno, ok := err.(*net.OpError); ok && errno.Err == syscall.EADDRNOTAVAIL {
			return nil, fmt.Errorf("host %s not reachable: %w", host, err)
		}
		return nil, fmt.Errorf("cannot allocate port: %w", err)
	}
	// defer l.Close() // close immediately
	port := l.Addr().(*net.TCPAddr).Port
	return &Reserved{Listener: l, Port: port}, nil
}

func EnsurePortAvailable(host string, desiredPort int) (*Reserved, error) {
	// Trường hợp “auto” ngay từ đầu
	if desiredPort == 0 {
		return FindAvailablePort(host)
	}

	// Thử bind đúng cổng mong muốn để kiểm tra
	l, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(desiredPort)))
	if err == nil {
		// l.Close() // OK – cổng rảnh, trả về
		return &Reserved{Listener: l, Port: desiredPort}, nil
	}
	if l != nil {
		_ = l.Close()
	}

	// Nếu cổng bận hoặc không đủ quyền ➜ fallback cổng khác
	if errors.Is(err, syscall.EADDRINUSE) || errors.Is(err, syscall.EACCES) {
		return FindAvailablePort(host)
	}

	// Các lỗi khác (host không tồn tại, v.v.)
	return nil, fmt.Errorf("cannot bind desired port %d: %w", desiredPort, err)
}
