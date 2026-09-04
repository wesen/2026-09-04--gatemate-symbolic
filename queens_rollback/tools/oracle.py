"""Independent ordered eight-queens oracle; no domains or rollback records."""

def valid_board(rows):
    return (len(rows) == 8 and all(0 <= r < 8 for r in rows)
            and all(rows[a] != rows[b] and abs(rows[a]-rows[b]) != b-a
                    for a in range(8) for b in range(a+1,8)))


def solve_rows():
    result = []
    def visit(rows):
        col = len(rows)
        if col == 8:
            result.append(tuple(rows))
            return
        for row in range(8):
            if all(row != old and abs(row-old) != col-c for c,old in enumerate(rows)):
                visit(rows+[row])
    visit([])
    return result


def unpack(word):
    return tuple((word >> (3*c)) & 7 for c in range(8))
