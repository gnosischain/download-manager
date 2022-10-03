
# Download Manager

Utility CLI to facilitate downloads of big files in chunks. This CLI is still in alpha phase and experimental.

# How to use it

Download the CLI from `multipart` branch

- `curl https://raw.githubusercontent.com/gnosischain/download-manager/main/releases/linux/amd64/download-manager --output ./download-manager`

Make it executable

- `chmod +x ./download-manager`

If not output path is passed, file will be downloaded in current working directory.

- `download-manager fetch -u {https://url-to-file} -f {filename}`

Otherwise specify an output path explicitly.

- `download-manager fetch -u {https://url-to-file} -f {filename} -o {output-path}`

Resume download from specific part

- `download-manager fetch -u {https://url-to-file} -f {filename} -p {part-number}`
