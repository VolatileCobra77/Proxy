Functions:
  This project will forward traffic from a pc with the client installed to my coordination servers, who will then complete the request, and respond with the data

Plan:
  Client:
     1. Accept local connections
        - HTTP CONNECT
        - SOCKS5
     2. Open a tunnel to the server
        - Send destination info
        - Wait for confirm
     3. Pipe
        - Clean shutdown
        - No Leaks