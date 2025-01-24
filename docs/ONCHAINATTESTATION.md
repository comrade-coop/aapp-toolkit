# On-Chain Attestation

## Context 🔧
Often, aApp applications act as agents natively integrated with blockchain technologies. This means they rely on blockchain wallets, which opens up the opportunity for the aApp instance to manage the DNS zone using blockchain technology.

There are two primary approaches for DNS zone management:

1. **Fully Blockchain-Managed DNS Zones**: Utilizing technologies like [Unstoppable Domains](https://unstoppabledomains.com/), the entire DNS zone can be managed directly on the blockchain.
2. **Hybrid Approach**: The DNS zone is managed off-chain, but the DNS challenge related to the certificate issuance process is attested on-chain. This ensures the TLS certificate can be matched with the on-chain attested certificate request. 

This document focuses on the hybrid approach.

---

## Hybrid On-Chain Attestation Process 📊

### Step-by-Step Process 📘

1. **aApp Initialization** 🛠️
   - During the initial bootstrapping, the aApp instance is seeded with a blockchain wallet and private key as part of the aApp manifest.

2. **Bootstrapping Process** 🛡️
   - **2.1.** A Certificate Signing Request (CSR) is prepared and submitted to the certificate authority (e.g., Let's Encrypt).
   - **2.2.** The DNS challenge is signed using the private key associated with the CSR.
   - **2.3.** Trusted Execution Environment (TEE) quotes are generated, including the signed DNS challenge and public key.
   - **2.4.** An attestation report is submitted on-chain using Automata Network (Automata DCAP Attestation On-Chain).

3. **DNS Record Update** 🔀
   - The developer (DNS zone owner) validates the on-chain attestation and updates the DNS record with the challenge.

---

## Public Transparency and Validation 📃

1. **End-User Validation** 🔓
   - End users can validate the certificate used in the TLS connection and ensure the public key matches the on-chain attestation.

2. **Community Transparency** 📚
   - The community can leverage the Certificate Transparency Log and Automata On-Chain Attestation to observe, match, and correlate the consistency between the certificate in use and the attested CSRs.

---

## Further Work 💡

1. **Full On-Chain DNS Record Management**
   - Implementing full on-chain DNS record management would enhance transparency by explicitly recording DNS record tampering rather than just detecting mismatches.

2. **Custom Certificate Attributes**
   - Using a certificate authority that allows custom (non-whitelisted) attributes in the certificate could simplify end-user validation.

---

For more information, refer to [aApp Toolkit Documentation](../README.md).