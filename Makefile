root	:=		$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: clean build deploy publish

clean:
	rm -rfv bin

build:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/main

deploy:
	sam package --output-template-file "${root}"/packaged.yml --s3-bucket "${bucket}"
	sam deploy --stack-name "${stack}" --capabilities CAPABILITY_IAM --template-file "${root}/packaged.yml"

publish:
	sam package --output-template-file "${root}"/packaged.yml --s3-bucket "${bucket}"
	sam publish --template "${root}/packaged.yml" --region "${AWS_DEFAULT_REGION}"
