FROM python:3.9-slim

# Install SSH server and other utilities
RUN apt-get update && apt-get install -y \
    openssh-server \
    curl \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*

# Create SSH directory and set up SSH
RUN mkdir /var/run/sshd
RUN echo 'root:testpassword' | chpasswd
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config

# SSH login fix
RUN sed 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' -i /etc/pam.d/sshd

# Create test user
RUN useradd -m -s /bin/bash testuser && echo 'testuser:testpass' | chpasswd

# Copy health check server
COPY health_server.py /app/health_server.py
COPY start.sh /app/start.sh
RUN chmod +x /app/start.sh

# Expose ports
EXPOSE 22 8080

# Set working directory
WORKDIR /app

# Start services
CMD ["/app/start.sh"]
