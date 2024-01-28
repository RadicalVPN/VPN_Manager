REMOTEHOST?=foobar
TARGET_OS=linux
TARGET_ARCH=amd64

deploy:
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -o vpn-manager
	rsync -avzp -e ssh --exclude='*.env' --exclude='*.git' ./vpn-manager root@$(REMOTEHOST):/root/VPN_Manager/