class MonacoEditor extends HTMLElement {
    constructor() {
      super();
      this.attachShadow({ mode: 'open' }); // Create a shadow DOM
    }
  
    connectedCallback() {
      const shadow = this.shadowRoot;
  
      // Defer loading Monaco Editor until connected
      import('monaco-editor')
        .then(monaco => {
          const container = shadow.createElement('div');
          container.style.width = '100%';
          container.style.height = '300px'; // Adjust height as needed
  
          shadow.appendChild(container);
  
          // Create the editor instance
          this.editor = monaco.editor.create(container, {
            value: this.getAttribute('value') || '', // Set initial value
            language: this.getAttribute('language') || 'javascript',
            theme: 'vs-dark' // Choose a theme (see Monaco Editor docs)
          });
  
          // Handle value changes (optional)
          this.editor.onDidChangeModelContent(() => {
            const newContent = this.editor.getValue();
            this.dispatchEvent(new CustomEvent('codeChange', { detail: newContent }));
          });
        })
        .catch(error => console.error('Error loading Monaco Editor:', error));
    }
  
    // Optional methods to interact with the editor (if needed)
    setValue(value) {
      if (this.editor) {
        this.editor.setValue(value);
      }
    }
  
    getValue() {
      if (this.editor) {
        return this.editor.getValue();
      }
      return '';
    }
  }
  
  customElements.define('monaco-editor', MonacoEditor);
  