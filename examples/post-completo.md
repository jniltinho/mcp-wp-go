<p><img src="/wp-content/uploads/2026/09/capa-do-post.webp" alt="Capa do artigo" width="1440" height="810" /></p>

## Instalação

Este exemplo combina **Markdown** e HTML confiável para publicação no WordPress.

![Tela do aplicativo](/wp-content/uploads/2026/09/tela-do-aplicativo.webp "Aplicativo em execução")

## Script de instalação

```bash
set -euo pipefail
curl -fsSL "https://example.com/install.sh" -o install.sh
chmod +x install.sh
./install.sh
```

## Configuração

```yaml
service:
  port: "{{ app_port }}"
  enabled: true
```

```javascript
const message = "<conteudo seguro>";
console.log(message);
```

## Vídeo

<div class="video-container"><iframe src="https://www.youtube.com/embed/jNQXAC9IVRw" title="Vídeo no YouTube" loading="lazy" allowfullscreen></iframe></div>

<div class="wp-note">Blocos HTML devem vir somente de arquivos confiáveis.</div>
