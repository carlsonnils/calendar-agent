import { mount } from "svelte";
import "./styles/audioRecorder.css";
import "./styles/display.css";
import "./styles/layout.css";
import "./styles/theme.css";
import App from "./App.svelte";

const app = mount(App, {
    target: document.getElementById("app"),
});

export default app;
