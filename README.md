
# Download Manager

Utility CLI to facilitate downloads of big files in chunks.

# How to use it

Make it executable

- `chmod +x ./download-manager`

If not output path is passed, file will be downloaded in current working directory.

- `download-manager fetch -u {https://url-to-file} -f {filename}`

Otherwise specify an output path explicitly.

- `download-manager fetch -u {https://url-to-file} -f {filename} -o {output-path}`

