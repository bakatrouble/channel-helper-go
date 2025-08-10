#[unsafe(no_mangle)]
pub extern "C" fn hash_image(buf: *const u8, size: u32) -> *mut std::ffi::c_char {
    let image_bytes = unsafe { std::slice::from_raw_parts(buf, size as usize) };
    let img = image::load_from_memory(image_bytes)
        .expect("Failed to load image from bytes");
    let hasher = imagehash::PerceptualHash::new()
        .with_image_size(8, 8)
        .with_hash_size(8, 8);
    std::ffi::CString::new(hasher.hash(&img).to_string()).unwrap().into_raw()
}
