from dataclasses import dataclass
from typing import List, Set, Dict
from collections import defaultdict

@dataclass
class Production:
    lhs: str
    rhs: List[str]

class LALRGrammarOptimizer:
    def __init__(self):
        self.productions: List[Production] = []
        self.nonterminals: Set[str] = set()
        self.terminals: Set[str] = set()
        self.first_sets: Dict[str, Set[str]] = {}
        self.follow_sets: Dict[str, Set[str]] = {}

    def add_production(self, line: str):
        """Add a production rule from a string in the format 'LHS -> RHS'"""
        parts = line.split('->')
        lhs = parts[0].strip()
        self.nonterminals.add(lhs)
        
        # Handle multiple alternatives (|)
        rhs_alternatives = parts[1].split('|')
        for alt in rhs_alternatives:
            symbols = [s.strip() for s in alt.strip().split()]
            if symbols:  # Skip empty alternatives
                self.productions.append(Production(lhs, symbols))
                # Add terminals (lowercase or special characters)
                for symbol in symbols:
                    if not symbol[0].isupper() and symbol not in {':', ';', ',', '[', ']', '(', ')', '+', '-', '*', '/', '%', '=', '!=', '>', '<', '>=', '<='}:
                        self.terminals.add(symbol)

    def eliminate_left_recursion(self):
        """Eliminate direct and indirect left recursion"""
        ordered_nonterminals = list(self.nonterminals)
        new_productions = []
        
        for i, Ai in enumerate(ordered_nonterminals):
            # First eliminate indirect recursion
            productions_i = [p for p in self.productions if p.lhs == Ai]
            new_prods_i = []
            
            for prod in productions_i:
                if prod.rhs and prod.rhs[0] in ordered_nonterminals[:i]:
                    # Replace Aj → γ with current production
                    aj_productions = [p for p in self.productions if p.lhs == prod.rhs[0]]
                    for aj_prod in aj_productions:
                        new_rhs = aj_prod.rhs + prod.rhs[1:]
                        new_prods_i.append(Production(Ai, new_rhs))
                else:
                    new_prods_i.append(prod)
            
            # Then eliminate direct recursion
            alpha_prods = []  # Productions with direct left recursion
            beta_prods = []   # Productions without direct left recursion
            
            for prod in new_prods_i:
                if prod.rhs and prod.rhs[0] == Ai:
                    alpha_prods.append(prod.rhs[1:])
                else:
                    beta_prods.append(prod.rhs)
            
            if alpha_prods:  # If there is direct left recursion
                new_nt = f"{Ai}'"
                self.nonterminals.add(new_nt)
                
                # Add new productions for Ai
                for beta in beta_prods:
                    if beta:
                        new_productions.append(Production(Ai, beta + [new_nt]))
                    else:
                        new_productions.append(Production(Ai, [new_nt]))
                
                # Add new productions for Ai'
                for alpha in alpha_prods:
                    if alpha:
                        new_productions.append(Production(new_nt, alpha + [new_nt]))
                new_productions.append(Production(new_nt, ['ε']))  # Add epsilon production
            else:
                new_productions.extend(new_prods_i)
        
        self.productions = new_productions

    def left_factor(self):
        """Apply left factoring to the grammar"""
        changed = True
        while changed:
            changed = False
            new_productions = []
            processed_lhs = set()
            
            for lhs in self.nonterminals:
                if lhs in processed_lhs:
                    continue
                    
                prods_with_lhs = [p for p in self.productions if p.lhs == lhs]
                if not prods_with_lhs:
                    continue
                
                # Group productions by common prefix
                prefix_groups = defaultdict(list)
                for prod in prods_with_lhs:
                    prefix = tuple(prod.rhs[:1])  # Take first symbol as prefix
                    prefix_groups[prefix].append(prod.rhs)
                
                # Apply left factoring where needed
                for prefix, group in prefix_groups.items():
                    if len(group) > 1:  # Common prefix found
                        changed = True
                        new_nt = f"{lhs}_{len(new_productions)}"
                        self.nonterminals.add(new_nt)
                        
                        # Add new production with common prefix
                        new_productions.append(Production(lhs, list(prefix) + [new_nt]))
                        
                        # Add productions for the rest
                        for rhs in group:
                            rest = rhs[len(prefix):]
                            if rest:
                                new_productions.append(Production(new_nt, rest))
                            else:
                                new_productions.append(Production(new_nt, ['ε']))
                    else:
                        new_productions.append(Production(lhs, group[0]))
                
                processed_lhs.add(lhs)
            
            # Add remaining productions
            new_productions.extend([p for p in self.productions if p.lhs not in processed_lhs])
            self.productions = new_productions

    def optimize(self):
        """Apply all optimization techniques"""
        self.eliminate_left_recursion()
        self.left_factor()
        self.remove_empty_productions()
        self.remove_unit_productions()

    def remove_empty_productions(self):
        """Remove ε-productions where possible"""
        nullable = set()
        for prod in self.productions:
            if 'ε' in prod.rhs:
                nullable.add(prod.lhs)
        
        changed = True
        while changed:
            changed = False
            for prod in self.productions:
                if prod.lhs not in nullable:
                    if all(symbol in nullable for symbol in prod.rhs):
                        nullable.add(prod.lhs)
                        changed = True
        
        new_productions = []
        for prod in self.productions:
            if 'ε' not in prod.rhs:
                new_productions.append(prod)
            else:
                # Add alternative productions without epsilon
                nullable_positions = [i for i, symbol in enumerate(prod.rhs) if symbol in nullable]
                for i in range(len(nullable_positions) + 1):
                    for combination in self._combinations(nullable_positions, i):
                        new_rhs = [s for j, s in enumerate(prod.rhs) if j not in combination]
                        if new_rhs:
                            new_productions.append(Production(prod.lhs, new_rhs))
        
        self.productions = new_productions

    def remove_unit_productions(self):
        """Remove unit productions (A → B)"""
        unit_productions = {}
        for prod in self.productions:
            if len(prod.rhs) == 1 and prod.rhs[0] in self.nonterminals:
                if prod.lhs not in unit_productions:
                    unit_productions[prod.lhs] = set()
                unit_productions[prod.lhs].add(prod.rhs[0])
        
        # Compute transitive closure
        for k in self.nonterminals:
            for i in self.nonterminals:
                if i in unit_productions:
                    for j in self.nonterminals:
                        if j in unit_productions.get(k, set()):
                            unit_productions[i] = unit_productions.get(i, set()) | unit_productions.get(j, set())
        
        # Replace unit productions
        new_productions = []
        for prod in self.productions:
            if len(prod.rhs) != 1 or prod.rhs[0] not in self.nonterminals:
                new_productions.append(prod)
                if prod.lhs in unit_productions:
                    for unit in unit_productions[prod.lhs]:
                        new_productions.append(Production(unit, prod.rhs))
        
        self.productions = new_productions

    @staticmethod
    def _combinations(items, n):
        """Helper function to generate combinations"""
        if n == 0:
            yield []
            return
        for i in range(len(items) - n + 1):
            for cc in LALRGrammarOptimizer._combinations(items[i + 1:], n - 1):
                yield [items[i]] + cc

    def __str__(self):
        """Convert grammar to string representation"""
        result = []
        current_lhs = None
        current_productions = []
        
        for prod in sorted(self.productions, key=lambda p: p.lhs):
            if current_lhs != prod.lhs:
                if current_productions:
                    result.append(f"{current_lhs} -> {' | '.join(' '.join(rhs) for rhs in current_productions)}")
                current_lhs = prod.lhs
                current_productions = []
            current_productions.append(prod.rhs)
        
        if current_productions:
            result.append(f"{current_lhs} -> {' | '.join(' '.join(rhs) for rhs in current_productions)}")
        
        return '\n'.join(result)


optimizer = LALRGrammarOptimizer()

# Add your grammar rules
grammar_rules = """
program_all  -> procedures main

procedures   -> procedures PROCEDURE proc_head IS declarations BEGIN commands END
             | procedures PROCEDURE proc_head IS BEGIN commands END
             | 

main         -> PROGRAM IS declarations BEGIN commands END
             | PROGRAM IS BEGIN commands END

commands     -> commands command
             | command

command      -> identifier := expression;
             | IF condition THEN commands ELSE commands ENDIF
             | IF condition THEN commands ENDIF
             | WHILE condition DO commands ENDWHILE
             | REPEAT commands UNTIL condition;
             | FOR pidentifier FROM value TO value DO commands ENDFOR
             | FOR pidentifier FROM value DOWNTO value DO commands ENDFOR
             | proc_call;
             | READ identifier;
             | WRITE value;

proc_head    -> pidentifier ( args_decl )

proc_call    -> pidentifier ( args )

declarations -> declarations, pidentifier
             | declarations, pidentifier[num:num]
             | pidentifier
             | pidentifier[num:num]

args_decl    -> args_decl, pidentifier
             | args_decl, T pidentifier
             | pidentifier
             | T pidentifier

args         -> args, pidentifier
             | pidentifier

expression   -> value
             | value + value
             | value - value
             | value * value
             | value / value
             | value % value

condition    -> value = value
             | value != value
             | value > value
             | value < value
             | value >= value
             | value <= value

value        -> num
             | identifier

identifier   -> pidentifier
             | pidentifier[pidentifier]
             | pidentifier[num]
"""

# Add each production
for line in grammar_rules.strip().split('\n'):
    if line and '->' in line:
        optimizer.add_production(line)

# Optimize the grammar
optimizer.optimize()

# Print the optimized grammar
print(optimizer)