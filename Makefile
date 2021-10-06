OUT = cb_image

all: clean build deploy seed

seed:
	awslocal s3 cp data/00000000-aaaa-1111-bbbb-abc123def456.png s3://checkboximageupload/00000000-aaaa-1111-bbbb-abc123def456/original.png

build:
	cd src; GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../bin/$(OUT)

deploy:
	terraform -chdir init
	terraform -chdir=infra apply -auto-approve

clean:
	rm -f bin/$(OUT)
	rm -f infra/$(OUT).zip
