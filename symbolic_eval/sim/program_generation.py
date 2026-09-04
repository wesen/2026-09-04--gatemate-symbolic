"""Generate bounded, executable fragments checked against the semantic machine."""
from opcodes import encode, Tag
from stack_model import Machine


def gen_program(rng, length=150, total_depth=8, illegal_p=0.06):
    words = []
    state = Machine(program=[], total_depth=total_depth)
    while len(words) < length:
        base = len(words)
        if rng.random() < illegal_p:
            # Every injected fault is validated rather than guessed from tags.
            faults = [encode(0x1f), encode(9), encode(3), encode(0x10)]
            rng.shuffle(faults)
            for word in faults:
                trial = Machine(program=words+[word], pc=base, total_depth=total_depth,
                                stack=list(state.stack))
                trial.step()
                if trial.fault is not None:
                    return words+[word], [v.tag for v in state.stack]
        fragments = [[encode(0, rng.randrange(-20, 21))], [encode(1)], [encode(2)]]
        fragments += [[encode(op)] for op in range(3, 11)]
        fragments += [[encode(0x0d)]]
        # Both paths join at a real next instruction, with no padding execution.
        fragments += [[encode(rng.choice([1, 2])), encode(0x0c, base+3), encode(0x0b, base+3)]]
        # Bounded nested calls: outer -> inner -> RET -> RET -> jump to join.
        fragments += [[encode(0x0f, base+2), encode(0x0b, base+5),
                       encode(0x0f, base+4), encode(0x10), encode(0x10)]]
        rng.shuffle(fragments)
        for fragment in fragments:
            if len(words)+len(fragment) > length:
                continue
            trial = Machine(program=words+fragment, pc=base, total_depth=total_depth,
                            stack=list(state.stack))
            for _ in range(12):
                if trial.pc == base+len(fragment) or trial.fault is not None:
                    break
                trial.step()
            if trial.fault is None and trial.pc == base+len(fragment) and not trial.rstack:
                words += fragment
                state = trial
                break
        else:
            # Only possible if the remaining word budget cannot fit a fragment.
            break
    words.append(encode(0x0e))
    return words, [v.tag for v in state.stack]
