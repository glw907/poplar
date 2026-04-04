-- Unwrap layout tables and divs: extract contents with paragraph
-- breaks between sections so marketing emails don't run together.

function Table(el)
  local blocks = {}
  if el.head and el.head.rows then
    for _, row in ipairs(el.head.rows) do
      for _, cell in ipairs(row.cells) do
        if #blocks > 0 then
          table.insert(blocks, pandoc.Para{})
        end
        for _, block in ipairs(cell.contents) do
          table.insert(blocks, block)
        end
      end
    end
  end
  for _, body in ipairs(el.bodies) do
    for _, row in ipairs(body.body) do
      for _, cell in ipairs(row.cells) do
        if #blocks > 0 then
          table.insert(blocks, pandoc.Para{})
        end
        for _, block in ipairs(cell.contents) do
          table.insert(blocks, block)
        end
      end
    end
  end
  return blocks
end

function Div(el)
  return el.content
end
