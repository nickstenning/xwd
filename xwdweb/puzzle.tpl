<h1>{{.Title}}</h1>
<h2>{{.Author}}</h2>

<table class="puzzle" cellspacing="0">
  {{range .Solution}}
  <tr>
    {{range .}}
      {{if .Black}}
        <td class="black"></td>
      {{else}}
        <td class="white i-{{index .Coords 0}} j-{{index .Coords 1}}">
          {{if isnumcell .}}
          <span class="number">{{inc .Num}}</span>
          {{end}}
          <input type="text" maxlength="1">
        </td>
      {{end}}
    {{end}}
  </tr>
  {{end}}
</table>

<h3>Across</h3>
<ul>
  {{range .CluesAcross}}
  <li class="across-{{inc .Num}}">
    <strong>{{inc .Num}}</strong>
    {{.Clue}}
  </li>
  {{end}}
</ul>

<h3>Down</h3>
<ul>
  {{range .CluesDown}}
  <li class="across-{{inc .Num}}">
    <strong>{{inc .Num}}</strong>
    {{.Clue}}
  </li>
  {{end}}
</ul>
