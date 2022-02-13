data "archive_file" "lambda" {
  type = "zip"

  source_dir  = "${path.module}/????"
  output_path = "${path.module}/coinbase.zip"
}
