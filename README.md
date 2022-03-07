"ascii-game-engine" is a set of tools for developing text based terminal games on Linux, with Go.  This is still very much a work in progress, but there are some included demos.  The key idea is a layer of abstraction for drawing colored text and "pixels" on a terminal window that goes beyond just a coordinate based "place a color here" system, moving towards updating in memory values and automatically updating the canvas to what it should be.  It supports "zones", or switchable places that a cursor can have a different location in, and allows for menus, different view contexts, and most importantly very quick sketches of terminal game ideas that a developer may have in mind.  Newer features involve a networked peer-to-peer system (moving towards one peer as a host) as well as support for audio effects.  Other cool features involve an ASCII "Doom-style" face icon that can be updated based on events.

Main Features:

 - High Performance Good Pathfinding ✅
 - Lightweight Curses Alternative ✅
 - Image to ASCII ✅
 - Frame Capped "GIF" Rendering in ASCII ✅
 - Dialogue Tree Editor
 - Topographic Map Generation
 - Symbolic Mapping for Maps
 - Concurrent Input Handling and Display ✅
 - No Dependencies ✅ (Linux only though)
 - Easy P2P networking over UDP ✅
 - Support for audio and sound effects ✅
 - Pixel based frame editor ✅ (Overhaul coming)
 - Entity selection and targeting
