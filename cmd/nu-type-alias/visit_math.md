# Tree

Let $N$ be the set of all nodes.

For a $n \in N$, let $C_{n} \subset N$.

> $C_{n}$ is the set of children of $n$.

Let $R \in N$

> $R$ is the root.

$P(n) \in \{p : n \in C_{p}\}$

> $P(n)$ is one such node which has $n$ as one of its children.

$\forall n \in (N - \{R\}) \left[\left|\left\{p : n \in C_{p}\right\}\right| = 1\right]$

> Each child $n$ has exactly one parent $p$, and thus one
> particular value for $P(n)$.

For $n \in C_{p}$, $I_{n} \in \mathbb{N}$ is the index of the node
$n$ in the children list.

$O \subset N$

> $O$ is the set of all nodes which are comments.

# Comment blocks

A particular block $B \subset N$ satisfies the following:

$\forall n \in B [n \in O]$

> All nodes in $B$ are comments.

$\forall a \in B \forall b \in B [a \neq b \to P(a) = P(b)]$

> All nodes in $B$ share the same parent.

$\exists ! m \in B \neg \exists n \in B [I_{m} = I_{n} + 1]$

> $m$ is the node with minimum index in $B$.

$\forall a \in (B - \{m\}) \exists ! b \in B [I_{a} = I_{b}+1]$

> The nodes in $B$ are consecutive.

# Requirements

If a block $B$ is followed with a child that is a:

- cmd: it should be considered a cmd comment
- var: it should be considered a var comment
- else: it should be considered a lone comment

