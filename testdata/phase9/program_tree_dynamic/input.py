class Node:
    def __init__(self, value):
        self.value = value
        self.left = None
        self.right = None

def insert(root, value):
    if root is None:
        return Node(value)
    if value < root.value:
        root.left = insert(root.left, value)
    elif value > root.value:
        root.right = insert(root.right, value)
    return root

def inorder(root):
    if root is None:
        return []
    return inorder(root.left) + [root.value] + inorder(root.right)

def height(root):
    return 0 if root is None else 1 + max(height(root.left), height(root.right))

def contains(root, value):
    while root is not None:
        if value == root.value:
            return True
        root = root.left if value < root.value else root.right
    return False

def level_order(root):
    out, queue = [], [root] if root else []
    while queue:
        node = queue.pop(0)
        out.append(node.value)
        queue.extend(n for n in (node.left, node.right) if n)
    return out

root = None
for v in (50, 30, 70, 20, 40, 60, 80, 30, 65):
    root = insert(root, v)
print(inorder(root), height(root), level_order(root))
print([contains(root, v) for v in (40, 65, 99, 50)], inorder(None), height(None))
words = None
for w in "the quick brown fox jumps over the lazy dog".split():
    words = insert(words, w)
print(inorder(words), height(words))
