## Getting Started

By default, this crate enables all its features - `configfs`, `tpm` and `legacy`.

### Features
- `tpm`: This feature must be used for a VM that only exports its attestation report via a TPM module. VMs that support this include Azure Confidential VMs.
- `configfs`: This feature should be the default for most Cloud Service Providers (CSP), such as AWS and GCP.
- `legacy`: This feature should only be enabled on CSPs that support this, and only on AMD SEV-SNP VMs. Do not use this where possible, as confidential computing development is moving to ConfigFS.

> We recommend leaving all features on unless you know which feature to exclude.

### Download Dependencies
When the `tpm` feature is enabled, the following dependencies must be installed:

```bash
sudo apt install pkg-config libtss2-dev
```

### Importing this crate
To use the crate as is, import it into your `Cargo.toml` as follows:

```toml
[dependencies]
coco-provider = "0.1.0"
```

To use this crate without specific features, eg. `tpm` feature, import it into your `Cargo.toml` as follows:

```toml
[dependencies]
coco-provider = { version = "0.1.0", default-features = false, features = ["configfs", "legacy"] }
```

### Using this Crate
This crate exports all the information you need in one function,`get_coco_provider`, which can be used as follows:

```rust
use coco_provider::get_coco_provider;

fn main() {
    let provider = get_coco_provider().unwrap();
    println!("Provider: {:?}", provider);
}
```