
# aApp Toolkit

The **aApp Toolkit** is a comprehensive set of tools designed to simplify building, deploying, and running applications on **Trusted Execution Environment (TEE)**-enabled hardware. Whether you're packaging your application, setting up deployment, or integrating essential services, the toolkit jumpstarts your journey into secure, TEE-based application development.

[![Discord](https://img.shields.io/badge/DISCORD-COMMUNITY-informational?style=for-the-badge&logo=discord)](https://discord.gg/fWwMSZdBF2)

## ⚠️ Disclaimer
This repository is currently under active development as we restructure and consolidate code from our other projects. Until the first alpha release is tagged, the main codebase may be unstable and subject to frequent changes. We recommend waiting for the first tagged release for a more stable version of the toolkit.

## Demo Video

[![Watch the video](https://img.youtube.com/vi/ZLCqmftx3gc/hqdefault.jpg)](https://youtu.be/ZLCqmftx3gc)

## 🚀 Features

- **Container-based Application Packaging**  
  Easily containerize your application for deployment in TEE environments.

- **TEE-Enabled Hardware Deployment**  
  Includes pre-configured base VM images for supported hardware platforms (Intel TDX and AMD SEV-SNP), ensuring a seamless setup process.

- **Composable Deployment Model**  
  Deploy your application in a single container-capable CVM or compose it within complex orchestration setups—ranging from confidential containers to confidential kubernetes clusters.

- **Confidential Networking**  
  Enable your application instances to form confidential networks for high availability, load balancing, and more.

- **Platform-Agnostic Services**  
  Build applications in any language or framework. The toolkit’s services follow the sidecar pattern and include:
  - **Remote Attestation**: Simplify verifying the integrity and authenticity of application instances.
  - **Reverse Proxy (Attestation-Aware)**: Manage ingress and egress traffic securely.
  - **Service Discovery**: Enable dynamic discovery of services across your confidential application cloud.
  - **Additional Tools**: Explore more features tailored for secure and efficient TEE application development.

## 🎯 Why Use aApp Toolkit?

Building applications for TEE-enabled hardware comes with unique challenges like secure packaging, attestation, and handling high availability. The **aApp Toolkit** abstracts these complexities, providing developers with streamlined, ready-to-use solutions. Focus on creating great applications without reinventing the wheel.

## 🛠️ Getting Started

Follow the steps below to start using the **aApp Toolkit**.

### Prerequisites

Ensure the following tools are installed:
- **Terraform**: [Install Terraform](https://developer.hashicorp.com/terraform/install)  
- **Mozilla SOPS**: [Install SOPS](https://github.com/getsops/sops/releases)

### Containerize Your Application

Package your application as a container image and publish it to a container registry. For open-source applications, consider using a reproducible container build process to enhance transparency and build trust with end users.

### Deploy Your Application

In the `terraform` folder, find examples for deploying container-enabled CVMs using public cloud confidential offerings like GCP and Azure. Then create your aApp manifest using `sops` and pass it as metadata to you CVM deployment script:

```yaml
apiVersion: alpha.aapp-toolkit.io/v1
kind: Application
spec:
  container:
    image: your-autonomous-app/awesomeapp:v1
  dns:
    zone: '*.your-autonomous-app.cloud'
    provider: 
      name: cloudflare
      env:
        - name: CF_API_KEY
          value: ENC[AES256_GCM,data:p673w==,iv:YY=,aad:UQ=,tag:A=]
        - name: CF_API_EMAIL
          value: ENC[AES256_GCM,data:CwE4O1s=,iv:2k=,aad:o=,tag:w==]
  ingress:
     rules:
      - http:
          paths:
          - path: "/web"
            backend:
              service:
                port:
                  number: 8080
  mtlsIngress:
     rules:
      - http:
          paths:
          - path: "/api-internal"
            backend:
              service:
                port:
                  number: 8080
sops:
# Metadata of your sops setup
```

### Bootstrap Your Application

Use the `aapp-toolkit-cli` to connect to deployed application instance, perform remote attestation, validate manifest and bootstrap it by passing encryption keys used in your `sops` setup.

### Test Your Application

Access the public ingress endpoint specified in the manifest to establish a secure connection to your aApp.

### Scale Out Your Application (Optional)

Deploy additional container-enabled CVMs using Terraform. Observe P2P-based discovery and the automatic bootstrapping of new aApp instances.

## 🧰 Under the Hood

The **aApp Toolkit** follows the sidecar design pattern. Services are delivered using proxies and operating system environment primitives (e.g., environment variables, `stdin`, and `stdout`). Built on leading open-source projects from the cloud-native ecosystem configured and extended for supporting confidential computing scenarios.

### High-Level Architecture

![High-level design](docs/assets/high-level-architecture.png)

### Core Concepts

- **P2P Bootstrapping**  
  The first instance is bootstrapped by the developer following a remote attestation process. The toolkit provides a minimal bootstrapping workflow that can be extended or integrated into a broader governance and transparency framework based on your application's or network's specific requirements. Subsequent instances are self-bootstrapped through mutual attestation. This process configures proxies, seeds initial secrets, and ensures traceability in the certificate-issuing process using a publicly auditable transparency log. 

- **DNS-Based Attested Service Discovery**  
  aApp instances are discoverable both publicly and privately using DNS. During the initial bootstrapping process, the application assumes ownership of a DNS zone, enabling seamless public and private communication. This DNS ownership plays a crucial role in facilitating the certificate issuance process during peer-to-peer (P2P) bootstrapping. Depending on the specific use case and the applicable governance and transparency framework, the DNS zone can be managed using blockchain technology through on-chain attestation. This approach ensures trust and accountability in the DNS zone's management. For more details, refer to the [On-Chain Attestation documentation](docs/ONCHAINATTESTATION.md).

- **High Availability**  
  Multiple aApp instances form a network, actively monitoring application state and reporting to service discovery components.

## 🏗️ Acknowledgements
We would like to express our gratitude to the [Constellation project](https://github.com/edgelesssys/constellation) for providing the build infrastructure used to create container-enabled confidential images and for their robust aTLS library, which powers the secure bootstrapping logic of the toolkit. Their contributions have been invaluable in advancing the development of secure and efficient TEE-based applications.  

We also extend our thanks to Microsoft for providing [Azure Attestation Services](https://azure.microsoft.com/en-us/products/azure-attestation) and [open-source libraries](https://github.com/Azure/confidential-computing-cvm-guest-attestation) that streamline Remote Attestation. Their support has been instrumental in enabling secure and efficient attestation workflows, further strengthening the foundation of confidential computing applications.

## 🤝 Contributing

This project is in its early stages. Significant contributions should be discussed with the team beforehand to ensure alignment with project goals. Join the conversation on [Discord](https://discord.gg/fWwMSZdBF2) or open/reply to issues to propose your ideas.

## 📄 License

This project is licensed under the [MIT License](LICENSE).
