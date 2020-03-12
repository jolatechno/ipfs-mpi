import sys

def install():
    pass
    #install package

def run(args):
    result = None
    print(result)

def uninstall():
    pass
    #uninstall package

if __name__ == "__main__":
    if sys.argv[0] == "install":
        install()

    if sys.argv[0] == "run":
        run(sys.argv[1:])

    if sys.argv[0] == "uninstall":
        uninstall()
