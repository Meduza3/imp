import sys

class VirtualMachine:
    def __init__(self):
        self.memory = [0] * 10350
        self.lasprint = self.memory
        self.instruction_pointer = 0
        self.instructions = []
        self.running = True

    def load_program(self, filename):
        with open(filename, 'r') as file:
            self.instructions = [line.strip() for line in file if line.strip() and not line.startswith('#')]

    def execute(self):
        co = 0
        while self.running and self.instruction_pointer < len(self.instructions):
            input("Press Enter to continue...")
            self.print_state()
            self.execute_instruction(self.instructions[self.instruction_pointer])

    def execute_instruction(self, instruction):
        parts = instruction.split()
        command = parts[0]
        if len(parts) > 1:
            argument = int(parts[1])
        else:
            argument = None

        if command == 'GET':
            self.memory[argument] = int(input(f"Enter value for memory[{argument}]: "))
            self.instruction_pointer += 1
        elif command == 'PUT':
            print(self.memory[argument])
            self.instruction_pointer += 1
        elif command == 'LOAD':
            self.memory[0] = self.memory[argument]
            self.instruction_pointer += 1
        elif command == 'STORE':
            self.memory[argument] = self.memory[0]
            self.instruction_pointer += 1
        elif command == 'LOADI':
            self.memory[0] = self.memory[self.memory[argument]]
            self.instruction_pointer += 1
        elif command == 'STOREI':
            self.memory[self.memory[argument]] = self.memory[0]
            self.instruction_pointer += 1
        elif command == 'ADD':
            self.memory[0] += self.memory[argument]
            self.instruction_pointer += 1
        elif command == 'SUB':
            self.memory[0] -= self.memory[argument]
            self.instruction_pointer += 1
        elif command == 'ADDI':
            self.memory[0] += self.memory[self.memory[argument]]
            self.instruction_pointer += 1
        elif command == 'SUBI':
            self.memory[0] -= self.memory[self.memory[argument]]
            self.instruction_pointer += 1
        elif command == 'SET':
            self.memory[0] = argument
            self.instruction_pointer += 1
        elif command == 'HALF':
            self.memory[0] //= 2
            self.instruction_pointer += 1
        elif command == 'JUMP':
            self.instruction_pointer += argument
        elif command == 'JPOS':
            if self.memory[0] > 0:
                self.instruction_pointer += argument
            else:
                self.instruction_pointer += 1
        elif command == 'JZERO':
            if self.memory[0] == 0:
                self.instruction_pointer += argument
            else:
                self.instruction_pointer += 1
        elif command == 'JNEG':
            if self.memory[0] < 0:
                self.instruction_pointer += argument
            else:
                self.instruction_pointer += 1
        elif command == 'RTRN':
            self.instruction_pointer = self.memory[argument]
        elif command == 'HALT':
            self.running = False
        else:
            raise ValueError(f"Unknown command: {command}")

    def print_state(self):
        print("Memory State:")
        for i, value in enumerate(self.memory):
            if value != 0:
                # Optionally, indicate if the value changed since the last print
                if value != self.lasprint[i]:
                    print(f"Memory[{i}]: {value} <---")
                else:
                    print(f"Memory[{i}]: {value}")
        print(f"Instruction Pointer: {self.instruction_pointer}")
        # Update lasprint with a copy of the current memory state
        self.lasprint = self.memory.copy()

if __name__ == "__main__":
    vm = VirtualMachine()
    if len(sys.argv) != 2:
        print("Usage: python tester.py <program_file>")
        sys.exit(1)
    program_file = sys.argv[1]
    vm.load_program(program_file)
    vm.execute()
    vm.print_state()
