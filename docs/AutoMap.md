When a struct auto mapping to a database's table, the below table describes how they change to each other:

<table>
    <tr>
    <td>go type's kind
    </td>
    <td>value method</td>
    <td>xorm type
    </td>
    </tr>
    <tr>
 <td>implemented Conversion</td>
 <td>Conversion.ToDB / Conversion.FromDB</td>
 <td>Text</td>
 </tr>
 <tr>
 <td>int, int8, int16, int32, uint, uint8, uint16, uint32</td>
 <td></td>
 <td> Int </td>
 </tr>
 <tr>
 <td>int64, uint64</td><td></td><td>BigInt</td>
 </tr>
 <tr><td>float32</td><td></td><td>Float</td>
 </tr>
 <tr><td>float64</td><td></td><td>Double</td>
 </tr>
 <tr><td>complex64, complex128</td>
 <td>json.Marshal / json.UnMarshal</td>
 <td>Varchar(64)</td>
 </tr>
 <tr>
 <td>[]uint8</td><td></td><td>Blob</td>
 </tr>
 <tr>
 <td>array, slice, map except []uint8</td>
 <td>json.Marshal / json.UnMarshal</td>
 <td>Text</td>
 </tr>
 <tr>
 <td>bool</td><td>1 or 0</td><td>Bool</td>
 </tr>
 <tr>
 <td>string</td><td></td><td>Varchar(255)</td>
 </tr>
 <tr>
 <td>time.Time</td><td></td><td>DateTime</td>
 </tr>
  <tr>
 <td>cascade struct</td><td>primary key field value</td><td>BigInt</td>
 </tr>
 <tr>
 <tr>
 <td>struct</td><td>json.Marshal / json.UnMarshal</td><td>Text</td>
 </tr>
 <tr>
 <td>
 Others
 </td>
 <td></td>
 <td>
 Text
 </td>
 </tr>
 </table>