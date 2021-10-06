provider "aws" {
  region = "us-east-1"
  access_key = "test"
  secret_key = "test"
  skip_credentials_validation = true
  skip_metadata_api_check = true
  skip_requesting_account_id = true
  s3_force_path_style = true
  endpoints {
    lambda = "http://127.0.0.1:4566"
    s3 = "http://127.0.0.1:4566"
  }
}

resource "aws_s3_bucket" "b" {
  bucket = "checkboximageupload"
  acl = "public-read"
}

data "archive_file" "cb_image" {
  type = "zip"
  source_file = "../bin/cb_image"
  output_path = "cb_image.zip"
}

resource "aws_lambda_function" "image_handler" {
  function_name = "cb_image"
  filename = "cb_image.zip"
  source_code_hash = "${data.archive_file.cb_image.output_base64sha256}"
  handler = "cb_image"
  role = "test_role"
  runtime = "go1.x"
  timeout = 5
  memory_size = 128
}
