
# Download Manager

Utility CLI to facilitate downloads of big files in chunks.

# How to use it

Download the CLI from `main` branch (does not support resume download from part)

- `curl https://raw.githubusercontent.com/gnosischain/download-manager/main/releases/linux/amd64/download-manager --output ./download-manager`

[Experimental] Download the CLI from `multipart` branch (does not support resume download from part)

- `curl https://raw.githubusercontent.com/gnosischain/download-manager/multipart/releases/linux/amd64/download-manager --output ./download-manager`

Make it executable

- `chmod +x ./download-manager`

If not output path is passed, file will be downloaded in current working directory.

- `download-manager fetch -u {https://url-to-file} -f {filename}`

Otherwise specify an output path explicitly.

- `download-manager fetch -u {https://url-to-file} -f {filename} -o {output-path}`
