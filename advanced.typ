#set page(
  paper: "a4",
  margin: (x: 1.8cm, y: 1.5cm),
)
#set text(
  font: "Noto Sans",
  size: 12pt,
)
#show heading: set block(above: 2.5em, below: 1.5em)

= Laptop Report for #q.ipAddress.answer

== A Heading

=== #q.QID1.text

#q.QID1.answer

== Some Lorem Ipsum

#lorem(30)

== A Table

#table(
  columns: (1fr, 1fr),
  [#q.startDate.text], [#q.startDate.answer],
  [#q.endDate.text], [#q.endDate.answer],
)
