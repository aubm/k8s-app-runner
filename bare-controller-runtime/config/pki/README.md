
```bash
# Generate certificate authority key and cert
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -subj "/CN=my-ca" -days 10000 -out ca.crt
openssl x509 -in ca.crt -out ca.crt.pem

# Generate certificate for webhooks server
openssl genrsa -out tls.key 2048
openssl req -new -key tls.key -subj "/CN=k8s-app-runner-webhook-service.k8s-app-runner-system.svc" -out tls.csr
openssl x509 -req -in tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt -days 10000

# optional, view the certificate
openssl x509 -noout -text -in ./tls.crt
```
