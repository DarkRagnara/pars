//pars is a simple framework for parser combinators. It is designed to be easy to use, yet powerful enough to allow solving real world problems.
//Parsers can be arranged into flexible stacks of more elemental parsers. Parser results can be transformed via transformers to easily convert their
//results into a more fitting format or to implement additional conditions that must be fulfilled for a successful parse.
//Complex parsers can be debugged by wrapping them with a logger.
//pars parsers can read from a string or from an io.Reader, so streaming parsing is possible if you need it.
package pars
