# Runs as an external program under a Port meta-process. One line in, one line
# out. fflush is not optional: without it awk buffers its output and the actor
# on the other side waits forever.
{
    n += 1
    total += $2
    printf "%d done, average %.1f ms\n", n, total / n
    fflush()
}
