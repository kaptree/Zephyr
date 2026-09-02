"""本地 mock OpenAI 兼容端点：POST 任意路径返回 200，用于 AI 配置连通性测试"""
from http.server import BaseHTTPRequestHandler, HTTPServer


class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(b'{"choices":[{"message":{"role":"assistant","content":"ok"}}]}')

    def log_message(self, *a):
        pass


if __name__ == "__main__":
    HTTPServer(("127.0.0.1", 18099), Handler).serve_forever()
