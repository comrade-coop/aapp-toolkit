use dcap_rs::types::quotes::version_4::QuoteV4;
use ethers::{
    prelude::*,
    providers::{Http, Provider},
    types::Bytes,
};
use std::env;
use std::sync::Arc;
use tdx::Tdx;

#[tokio::main]
async fn main() -> eyre::Result<()> {
    // Read provider URL from environment variable
    let provider_url = env::var("PROVIDER_URL")
        .expect("Environment variable PROVIDER_URL must be set");

    // Read wallet private key from environment variable
    let wallet_private_key = env::var("WALLET_PRIVATE_KEY")
        .expect("Environment variable WALLET_PRIVATE_KEY must be set");

    // Read contract address from environment variable
    let contract_address = env::var("CONTRACT_ADDRESS")
        .expect("Environment variable CONTRACT_ADDRESS must be set");

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

    // Connect to the Ethereum network
    let provider = Provider::<Http>::try_from(provider_url)?;
    let provider = Arc::new(provider);

    // Parse contract address
    let address = contract_address.parse::<Address>()
        .expect("Failed to parse CONTRACT_ADDRESS");

    // Contract ABI
    abigen!(
        Attestation,
        r#"[{
            "inputs": [
                {"internalType": "bytes", "name": "rawQuote", "type": "bytes"}
            ],
            "name": "verifyAndAttestOnChain",
            "outputs": [
                {"internalType": "bool", "name": "success", "type": "bool"},
                {"internalType": "bytes", "name": "output", "type": "bytes"}
            ],
            "stateMutability": "payable",
            "type": "function"
        }]"#
    );

    // Load wallet from private key
    let wallet = wallet_private_key
        .parse::<LocalWallet>()
        .expect("Failed to parse WALLET_PRIVATE_KEY")
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
                println!(
                    "Transaction mined in block: {}",
                    receipt.block_number.unwrap()
                );

                // Extract logs and print transaction result
                let logs = receipt.logs;
                println!("Logs: {:?}", logs);
            }
        }
        Err(e) => println!("Error sending transaction: {:?}", e),
    }

    Ok(())
}

