use rand::RngCore;

/// Generates 64 bytes of random data
/// Always guaranted to return something (ie, unwrap() can be safely called)
pub fn generate_random_data() -> Option<[u8; 64]> {
    let mut data = [0u8; 64];
    rand::thread_rng().fill_bytes(&mut data);
    Some(data)
}
