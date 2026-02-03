use std::collections::BTreeMap;
use std::fs;
use zellij_tile::prelude::*;

const DEFAULT_PATH: &str = "poc/status-line.txt";
const DEFAULT_INTERVAL_MS: u64 = 1000;

struct State {
    path: String,
    interval_ms: u64,
    line: String,
}

impl Default for State {
    fn default() -> Self {
        Self {
            path: DEFAULT_PATH.to_string(),
            interval_ms: DEFAULT_INTERVAL_MS,
            line: String::new(),
        }
    }
}

register_plugin!(State);

impl State {
    fn load_config(&mut self, configuration: BTreeMap<String, String>) {
        if let Some(path) = configuration.get("path") {
            if !path.is_empty() {
                self.path = path.to_string();
            }
        }
        if let Some(interval_ms) = configuration.get("interval_ms") {
            if let Ok(val) = interval_ms.parse::<u64>() {
                if val > 0 {
                    self.interval_ms = val;
                }
            }
        }
    }

    fn refresh_line(&mut self) {
        match fs::read_to_string(&self.path) {
            Ok(contents) => {
                let line = contents.lines().next().unwrap_or("");
                self.line = line.trim_end().to_string();
            }
            Err(_) => {
                self.line = format!("owlx: missing {}", self.path);
            }
        }
    }

    fn format_line(&self, cols: usize) -> String {
        let mut out = self.line.replace('\n', " ");
        if cols == 0 {
            return out;
        }
        let len = out.chars().count();
        if len > cols {
            out = out.chars().take(cols).collect();
        } else if len < cols {
            out.push_str(&" ".repeat(cols - len));
        }
        out
    }

    fn schedule_tick(&self) {
        let secs = (self.interval_ms as f64) / 1000.0;
        set_timeout(secs);
    }
}

impl ZellijPlugin for State {
    fn load(&mut self, configuration: BTreeMap<String, String>) {
        self.load_config(configuration);
        self.refresh_line();
        subscribe(&[EventType::Timer]);
        self.schedule_tick();
    }

    fn update(&mut self, event: Event) -> bool {
        if let Event::Timer(_) = event {
            self.refresh_line();
            self.schedule_tick();
            return true;
        }
        false
    }

    fn render(&mut self, _rows: usize, cols: usize) {
        let line = self.format_line(cols);
        print!("{}", line);
    }
}
