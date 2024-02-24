data "archive_file" "go_lambda" {
  type        = "zip"
  source_file = "../build/bootstrap"
  output_path = "../build/bootstrap.zip"
}

resource "aws_lambda_function" "main" {
  function_name    = "dyndb-test-lambda"
  role             = aws_iam_role.function_exec.arn
  runtime          = "provided.al2023"
  handler          = "bootstrap"
  memory_size      = 128
  timeout          = 28
  filename         = data.archive_file.go_lambda.output_path
  source_code_hash = data.archive_file.go_lambda.output_base64sha256
}

resource "aws_iam_role" "function_exec" {
  name               = "dyndb-test-function-exec"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_function_role.json
}

data "aws_iam_policy_document" "lambda_assume_function_role" {
  statement {
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role_policy_attachment" "function_basic_execution" {
  role       = aws_iam_role.function_exec.name
  policy_arn = data.aws_iam_policy.lambda_basic_execution.arn
}

data "aws_iam_policy" "lambda_basic_execution" {
  name = "AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "function_dynamodb_access_execution" {
  role       = aws_iam_role.function_exec.name
  policy_arn = data.aws_iam_policy.lambda_dynamodb_access_execution.arn
}

data "aws_iam_policy" "lambda_dynamodb_access_execution" {
  name = "AmazonDynamoDBFullAccess"
}
