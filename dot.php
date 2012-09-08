#!/usr/bin/env php
<?php
// Get the whole of stdin; let us hope it is not too big.
$buffer = '';
while ($l = fgets(STDIN)){
  $buffer .= $l;
}

$struct = json_decode($buffer);
printSruct($struct);

function printSruct($struct, $prefix = 'json'){
  if (is_object($struct)){
    echo "{$prefix} = {};\n";
  } elseif (is_array($struct)){
    echo "{$prefix} = [];\n";
  } else {
    echo "{$prefix} = ". json_encode($struct) .";\n";

    // No need to iterate if we already have a scalar
    return;
  }

  foreach ($struct as $k => $v){
    $k = json_encode($k);
    $outputKey = "{$prefix}[{$k}]";
    if (is_scalar($v)){
      $v = json_encode($v);
      echo "{$outputKey} = {$v};\n"; 
    } else {
      printSruct($v, $outputKey);
    }
  }
}

