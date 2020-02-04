// Find miss counts
var misses = g.V().In('miss-address').ToArray();
var counts = {}

for ( var i = 0; i < misses.length; i++ ) {
	// Ignore misses from outside TCC
	if ( misses[i].substring(0, 4) == '0x7f' ) {
		continue;
	}

	var missAddr = misses[i].substring(2);
	var missLoc = g.V(misses[i]).Out('miss-address').Limit(1).ToArray()[0];

	// Ignore instruction misses
	if ( missAddr == missLoc ) {
		continue;
	}

	while ( missAddr.length < 8 ) {
		missAddr = "0" + missAddr;
	}

	missAddr = "0x" + missAddr;

	// Make sure we have line data for this
	if ( g.V().Has("text-at-pc", missAddr).Count() == 0 ) {
		continue;
	}

	if ( !(missAddr in counts) ) {
		counts[missAddr] = 1;
	}
	else {
		counts[missAddr] += 1;
	}
}

// Take the line with the most misses
var maxAddr = undefined;
var maxCount = 0;

for ( var addr in counts ) {
	if ( maxAddr == undefined || counts[addr] > maxCount ) {
		maxAddr = addr;
		maxCount = counts[addr];
	}
}

if ( !maxAddr ) {
	throw new Error("No max address")
}

// Get line
var line = g.V().Has("text-at-pc", maxAddr).Limit(1).ToArray()[0];

if ( !line ) {
	throw new Error("Failed to find text for " + maxAddr)
}

g.Emit("Worst address: " +  maxAddr);
g.Emit("Worst number: " + line);
g.V(line).Out('line-content').All();