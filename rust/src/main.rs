use std::{env, error::Error, path::PathBuf};
use serde::Serialize;

#[derive(Serialize)]
struct ExtractedResult {
    text: String
}


fn main() -> Result<(), Box<dyn Error>> { 
    run()
}

fn run() -> Result<(), Box<dyn Error>> {
    let pdf_path = parse_pdf_path()?;
    let text = pdf_extract::extract_text(&pdf_path)?;

    let result = ExtractedResult { text };
    println!("{}", serde_json::to_string(&result)?);

    Ok(())
}

fn parse_pdf_path() -> Result<PathBuf, Box<dyn Error>> {
    let mut args = env::args_os();
    let _program = args.next();

    let Some(path) = args.next() else {
        return Err("missing PDF path".into());
    };

    if args.next().is_some() {
        return Err("expected exactly one PDF path".into());
    }

    let path = PathBuf::from(path);

    if !path.exists() {
        return Err(format!("file does not exist: {}", path.display()).into());
    }

    if !path.is_file() {
        return Err(format!("path is not a file: {}", path.display()).into());
    }

    Ok(path)
}
