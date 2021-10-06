# Solution

The solution I have come up with is reasonably simple. AWS already has many
services that provide the functionality that is required for this particular
task, so we will leverage them as much as possible

## Architecture Diagram

![Architecture Diagram](checkbox-architecture.png?raw=true "Architecture Diagram")

| Service| Description |
|--------|-------------|
| Front End | The front end application containing the new File Upload field is hosted somewhere. In this case Fargate, but it could be on EC2 or even static S3 |
| File Upload | This lambda service handles writing the file to S3 |
| Upload Field Metadata Cache | Configuration of the File Upload field, containing any restrictions on file type, size, image resolution etc |
| Upload Field Metadata Store | Configuration of the File Upload field, containing any restrictions on file type, size, image resolution etc |
| Live S3 Bucket | S3 Bucket where files will be stored if they meet the restrictions/criteria for the File Upload field |
| Audit S3 Bucket | S3 Bucket where files will be stored if the *do not* meet the restrictions/criteria for the File Upload field |
| File Metadata Store | Metadata store for successful and failed file uploads. Used for storing metadata such as who uploaded, when, where, what was wrong, etc |
| Image Reformatter | Service to reformat images in the requested format |
| Salesforce Integration | Service to integrate uploaded files in requested format |

### Flow

1. The Front End application containing the new File Upload field will send the upload request to the File Upload lambda
2. The File Upload lambda will check the Upload Field Metadata Cache (with the Upload Field Metadata Store as a failover) for the File Upload field configuration (allowed file types, size, etc)
3. Metadata on the uploaded file will be writted to the File Metadata Store. This may include things such as;
  - Who uploaded the file
  - Was the upload successfile
  - What was the actual file type
3. If the uploaded file meets the criteria, it will be written to the Live S3 bucket.
4. If the uploaded file does not meet the criteria, it will be written to the Audit S3 bucket for debugging/investigative purposes.
5. Uploads to either the Live or Audit S3 buckets will write the event to CloudWatch.
6. The Image Reformatter, Salesforce Integration, or Other file handler lambda wil read and write to the Live S3 bucket depending on implementation.


## Image Reformatter

The Image Reformatter is a service will return a reformatted image based on the request sent through.

### Request

```
{
  "uuid:" "00000000-aaaa-1111-bbbb-abc123def456",
  "format": "png",
  "resolution": {
	"width": 500,
	"height": 500
  }
}
```

### Response

```
{
  "url":"http://localhost:4566/checkboximageupload/00000000-aaaa-1111-bbbb-abc123def456/500x500.png"
}
```
### How it works

When the lambda is executed with a request it will handle it like so;

1. Check to see if we have a version of the image cached in S3 that conforms to the request
  - If so, then return the details of this image, otherwise continue
  - If there is no matching image in the cache we create it
2. Get the original image from S3. The image will reside at in a bucket with a key in the format of <UUID>/original.<file_ext>
  - If the request want the image in a format that doesn't match the original, e.g., our request wants `"format": "jpeg"`, but the original image is <UUID>/original.png, we will convert the original image to the desired format
3. Reformat the image to match the desired resolution. If either the width or the height of the requested resolution is 0, we will error out here.
4. Return a response containing the url for the image in the requested format and resolution.

### How to run it

You will need to install a few tools in order to run this locally;

- docker and docker-compose
- either awscli-local or awscli
- golang version 1.17

To get started, spin up localstack with `docker-compose up` using the provided docker-compose.yml configuration.

Once localstack is running you can compile the application, provision the infrastructure, and upload seed data using the command `make all`. This will provision and deploy the image reformatter lambda, as well as create a bucket `checkboximageupload` and seed an image with the key `00000000-aaaa-1111-bbbb-abc123def456/original.png`

To test the lambda, run the command `awslocal lambda invoke --function-name cb_image --payload '{"uuid": "00000000-aaaa-1111-bbbb-abc123def456", "format": "png", "resolution": {"width": 500, "height": 500}}' output.txt`. If you are using awscli instead of awscli-local, you will need to use the flag `--endpoint-url=http://localhost:4566` in order to issue the command against localstack.

Once complete, you should have a file `output.txt` containing the response from the lambda. If you still have localstack running, you should be able to fetch the requested image residing at the url contained in the response.

