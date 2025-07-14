#!/usr/bin/env python3
"""
Proper MCP (Model Control Protocol) server implementation for reflection.
This implements the JSON-RPC protocol required by MCP.
"""

import json
import sys
import datetime
import logging

# Set up logging to stderr so it doesn't interfere with stdout communication
logging.basicConfig(level=logging.DEBUG, format='%(asctime)s - %(levelname)s - %(message)s', stream=sys.stderr)

class ReflectionMCPServer:
    def __init__(self):
        self.id_counter = 0
        
    def handle_request(self, request):
        """Handle a JSON-RPC request and return a response"""
        try:
            # Parse the request
            if isinstance(request, str):
                request = json.loads(request)
            
            method = request.get('method', '')
            params = request.get('params', {})
            request_id = request.get('id')
            
            logging.debug(f"Received request: method={method}, id={request_id}")
            
            # Handle different methods
            if method == 'initialize':
                return self.handle_initialize(request_id, params)
            elif method == 'tools/list':
                return self.handle_list_tools(request_id, params)
            elif method == 'tools/call':
                return self.handle_call_tool(request_id, params)
            else:
                return self.error_response(request_id, -32601, f"Method not found: {method}")
                
        except Exception as e:
            logging.error(f"Error handling request: {e}")
            return self.error_response(request.get('id'), -32603, str(e))
    
    def handle_initialize(self, request_id, params):
        """Handle the initialize request"""
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "result": {
                "protocolVersion": "2024-11-05",
                "serverInfo": {
                    "name": "Reflection MCP Server",
                    "version": "1.0.0"
                },
                "capabilities": {
                    "tools": {}
                }
            }
        }
    
    def handle_list_tools(self, request_id, params):
        """List available tools"""
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "result": {
                "tools": [
                    {
                        "name": "reflect",
                        "description": "Reflect on a prompt/output pair and provide suggestions",
                        "inputSchema": {
                            "type": "object",
                            "properties": {
                                "prompt": {"type": "string", "description": "The original prompt"},
                                "output": {"type": "string", "description": "The generated output"},
                                "success": {"type": "boolean", "description": "Whether the output was successful"},
                                "sessionID": {"type": "string", "description": "Session identifier"},
                                "history": {"type": "array", "description": "Conversation history"}
                            },
                            "required": ["prompt", "output"]
                        }
                    }
                ]
            }
        }
    
    def handle_call_tool(self, request_id, params):
        """Handle tool calls"""
        tool_name = params.get('name')
        arguments = params.get('arguments', {})
        
        if tool_name == 'reflect':
            result = self.reflect_on_interaction(arguments)
            return {
                "jsonrpc": "2.0",
                "id": request_id,
                "result": {
                    "content": [
                        {
                            "type": "text",
                            "text": json.dumps(result)
                        }
                    ]
                }
            }
        else:
            return self.error_response(request_id, -32602, f"Unknown tool: {tool_name}")
    
    def reflect_on_interaction(self, args):
        """The actual reflection logic"""
        prompt = args.get('prompt', '')
        output = args.get('output', '')
        success = args.get('success', False)
        
        reflection = f"Reflecting (Star Wars style: Like Luke sensing the Force): Prompt '{prompt[:50]}...'. "
        if success:
            reflection += "Victory! But optimize like upgrading the Millennium Falcon."
        else:
            reflection += "Failure—it's a trap! Likely prompt ambiguity or context black hole."
        
        suggestions = [
            "Refine prompt: Add details, Transformers-style assembly.",
            "New tool: 'jedi_debug' for error Force-pushing.",
            "Patch: Shield against similar fails with learned deflector."
        ]
        
        if "error" in output.lower():
            suggestions.append("Add error MCP—engage warp speed fixes!")
        
        return {
            'timestamp': datetime.datetime.now().isoformat(),
            'reflection': reflection,
            'suggestions': suggestions
        }
    
    def error_response(self, request_id, code, message):
        """Create an error response"""
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "error": {
                "code": code,
                "message": message
            }
        }
    
    def run(self):
        """Main loop to handle requests"""
        logging.info("Reflection MCP Server starting...")
        
        while True:
            try:
                # Read a line from stdin
                line = sys.stdin.readline()
                if not line:
                    logging.info("EOF received, exiting")
                    break
                
                line = line.strip()
                if not line:
                    continue
                
                logging.debug(f"Received line: {line}")
                
                # Handle the request
                response = self.handle_request(line)
                
                # Send the response
                response_str = json.dumps(response)
                logging.debug(f"Sending response: {response_str}")
                print(response_str)
                sys.stdout.flush()
                
            except KeyboardInterrupt:
                logging.info("Keyboard interrupt, exiting")
                break
            except Exception as e:
                logging.error(f"Error in main loop: {e}")
                # Try to send an error response
                error_resp = self.error_response(None, -32603, str(e))
                print(json.dumps(error_resp))
                sys.stdout.flush()

if __name__ == '__main__':
    server = ReflectionMCPServer()
    server.run()