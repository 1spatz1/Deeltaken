// Define Monaco Editor web component
class MonacoEditor extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        const template = document.createElement('template');
        template.innerHTML = `
            <style>
            html, body, #container {
                position: absolute;
                left: 0;
                top: 0;
                width: 100%;
                height: 100%;
                margin: 0;
                padding: 0;
                overflow: hidden;
            }
            </style>
            <div id="container"></div>
        `;
        this.shadowRoot.appendChild(template.content.cloneNode(true));
    }

    connectedCallback() {
        const editorContainer = this.shadowRoot.getElementById('container');
    
        // Load Monaco Editor script dynamically
        const script = document.createElement('script');
        script.src = 'https://unpkg.com/monaco-editor@latest/min/vs/loader';
        script.onload = () => {
            // Configure loader
            window.require.config({ paths: { 'vs': 'https://unpkg.com/monaco-editor@latest/min/vs' }});
            // Set up Monaco Environment
            window.MonacoEnvironment = { getWorkerUrl: () => this.proxy };
    
            this.proxy = URL.createObjectURL(new Blob([`
                self.MonacoEnvironment = {
                    baseUrl: 'https://unpkg.com/monaco-editor@latest/min/'
                };
                importScripts('https://unpkg.com/monaco-editor@latest/min/vs/base/worker/workerMain.js');
            `], { type: 'text/javascript' }));
    
            // Load Monaco Editor and create an editor instance
            window.require(["vs/editor/editor.main"], () => {
                this.editor = monaco.editor.create(editorContainer, {
                    value: this.getAttribute('value'),
                    language: this.getAttribute('language'),
                    theme: this.getAttribute('theme')
                });
            });
        };
    
        document.head.appendChild(script);
    }
    

    disconnectedCallback() {
        if (this.editor) {
            this.editor.dispose();
            this.editor = null;
        }
        if (this.proxy) {
            URL.revokeObjectURL(this.proxy);
        }
    }
}

customElements.define('monaco-editor', MonacoEditor);