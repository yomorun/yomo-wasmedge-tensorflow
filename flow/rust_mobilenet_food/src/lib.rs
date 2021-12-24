use std::time::Instant;
use wasmedge_tensorflow_interface;
use wasmedge_bindgen::*;
use wasmedge_bindgen_macro::*;

#[wasmedge_bindgen]
pub fn infer(image_data: Vec<u8>) -> Result<Vec<u8>, String> {
    let start = Instant::now();

    let model_data: &[u8] = include_bytes!("lite-model_aiy_vision_classifier_food_V1_1.tflite");
    let labels = include_str!("aiy_food_V1_labelmap.txt");

    let flat_img = wasmedge_tensorflow_interface::load_jpg_image_to_rgb8(&image_data[..], 192, 192);
    println!("RUST: Loaded image in ... {:?}", start.elapsed());

    let mut session = wasmedge_tensorflow_interface::Session::new(&model_data, wasmedge_tensorflow_interface::ModelType::TensorFlowLite);
    session.add_input("input", &flat_img, &[1, 192, 192, 3])
           .run();
    let res_vec: Vec<u8> = session.get_output("MobilenetV1/Predictions/Softmax");

    let mut i = 0;
    let mut max_index: i32 = -1;
    let mut max_value: u8 = 0;
    while i < res_vec.len() {
        let cur = res_vec[i];
        if cur > max_value {
            max_value = cur;
            max_index = i as i32;
        }
        i += 1;
    }
    println!("RUST: index {}, prob {}", max_index, max_value);

    let confidence: String;
    if max_value > 200 {
        confidence = "is very likely".to_string();
    } else if max_value > 125 {
        confidence = "is likely".to_string();
    } else {
        confidence = "could be".to_string();
    }

    let ret_str: String;
    if max_value > 50 {
        let mut label_lines = labels.lines();
        for _i in 0..max_index {
            label_lines.next();
        }
        let food_name = label_lines.next().unwrap().to_string();
        ret_str = format!(
            "It {} a <a href='https://www.google.com/search?q={}'>{}</a> in the picture",
            confidence, food_name, food_name
        );
    } else {
        ret_str = "It does not appears to be a food item in the picture.".to_string();
    }

    println!(
        "RUST: Finished post-processing in ... {:?}",
        start.elapsed()
    );
    return Ok(ret_str.as_bytes().to_vec());
}
