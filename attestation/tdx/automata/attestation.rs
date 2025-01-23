use dcap_rs::types::quotes::version_4::QuoteV4;
use ethers::{
    prelude::*,
    providers::{Http, Provider},
    types::Bytes,
};
use std::sync::Arc;
use tdx::Tdx;

#[tokio::main]
async fn main() -> eyre::Result<()> {
    // Initialise a TDX object
    let tdx = Tdx::new();

    // Retrieve an attestation report with default options passed to the hardware device
    let raw_report = tdx.get_attestation_report_raw().unwrap();
    let report = QuoteV4::from_bytes(&raw_report);
    println!(
        "Attestation Report raw bytes: 0x{}",
        hex::encode(&raw_report)
    );
    println!("Attestation Report : {:?}", report);

    // Connect to Sepolia network
    let provider = Provider::<Http>::try_from(
        "https://sepolia.infura.io/v3/YOUR-PROJECT-ID"
    )?;
    let provider = Arc::new(provider);

    // Contract address
    let address = "0xE28ea4E574871CA6A4331d6692bd3DD602Fb4f76"
        .parse::<Address>()?;

    // Contract ABI
    abigen!(
        Attestation,
        r#"[
            function verifyAndAttestOnChain(bytes calldata rawQuote) 
                external 
                payable 
                returns (bool success, bytes memory output)
            event AttestationVerified(
                address indexed attester,
                bool success,
                bytes output
            )
            event VerificationFailed(
                address indexed attester,
                string reason
            )
        ]"#
    );

    // Load wallet from private key
    let wallet = "YOUR_PRIVATE_KEY"
        .parse::<LocalWallet>()?
        .with_chain_id(Chain::Sepolia);

    // Create contract instance with signer
    let client = SignerMiddleware::new(provider, wallet);
    let client = Arc::new(client);
    let contract = Attestation::new(address, client);

    // Call verifyAndAttestOnChain with payment
    let quote_bytes = Bytes::from(raw_report);
    let tx = contract
        .verify_and_attest_on_chain(quote_bytes)
        .value(U256::from(1_000_000_000_000_000u64)); // 0.001 ETH as example fee

    match tx.send().await {
        Ok(tx_receipt) => {
            println!("Transaction sent: {:?}", tx_receipt.tx_hash());
            if let Some(receipt) = tx_receipt.await? {
                println!("Transaction mined in block: {}", receipt.block_number.unwrap());
                
                // Parse the events from the receipt
                let events = contract.verify_and_attest_on_chain_filter().from_receipt(&receipt);
                match events {
                    Ok(logs) => {
                        for log in logs {
                            println!("Attestation verified for attester: {:?}", log.attester);
                            println!("Success: {:?}", log.success);
                            println!("Output: {:?}", log.output);
                        }
                    }
                    Err(e) => println!("Error parsing events: {:?}", e),
                }
            }
        }
        Err(e) => println!("Error sending transaction: {:?}", e),
    }

    Ok(())
}
