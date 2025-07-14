import json
import sys
import datetime

def reflect_iteration(input_data):
    try:
        data = json.loads(input_data)
        prompt = data.get('prompt', '')
        output = data.get('output', '')
        success = data.get('success', False)
        
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
        
        result = {
            'timestamp': datetime.datetime.now().isoformat(),
            'reflection': reflection,
            'suggestions': suggestions
        }
        return json.dumps(result)
    except Exception as e:
        return json.dumps({'error': str(e), 'reflection': 'Hang detected—like frozen in carbonite. Debug stdin loop.'})

if __name__ == '__main__':
    while True:
        try:
            line = sys.stdin.readline().strip()
            if not line:
                break  # Exit on empty/EOF to prevent hang
            response = reflect_iteration(line)
            print(response)
            sys.stdout.flush()
        except Exception as e:
            print(json.dumps({'error': str(e)}))
            sys.stdout.flush()
            break