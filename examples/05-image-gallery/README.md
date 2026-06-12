# Image Gallery Example

This example demonstrates listing image files from a directory using the widget mode table feature.

## Features

- **File scanning** - Automatically scans `pics/` directory for image files
- **Image preview** - Displays thumbnail previews in the table
- **Dynamic loading** - Images loaded from filesystem
- **Refresh capability** - Rescan directory for new images
- **Action buttons** - Open button for each image (extensible)

## Running the Example

```bash
cd examples/05-image-gallery
go run main.go
```

Or from the repository root:
```bash
go run ./examples/05-image-gallery
```

## Setup

### Add Images

1. Add your image files to the `pics/` subdirectory:
   ```bash
   cp ~/Pictures/*.jpg pics/
   cp ~/Pictures/*.png pics/
   ```

2. Or download sample images:
   ```bash
   wget -O pics/sample1.jpg https://picsum.photos/200/150
   wget -O pics/sample2.jpg https://picsum.photos/200/150
   wget -O pics/sample3.jpg https://picsum.photos/200/150
   ```

### Supported Formats

- PNG (.png)
- JPEG (.jpg, .jpeg)
- GIF (.gif)

## How It Works

### Scanning Files

```js
let scanImages = () => {
    let images = []
    let entries = os.readdir("pics")
    
    entries.each((filename) => {
        let lower = filename.lower()
        if (lower.ends_with(".png") || lower.ends_with(".jpg")) {
            images = images.append({
                filename: filename,
                path: "pics/" + filename
            })
        }
    })
    return images
}
```

### Loading Images in Table

```js
table.CreateCell((col, row) => {
    if (col == 1) {
        // Preview column: create image widget with first image as placeholder
        if (len(state.images) > 0) {
            let firstImage = state.images[0]
            let absPath = os.getwd() + "/" + firstImage.path
            return canvas.NewImageFromURI("file://" + absPath)
        }
        return widget.NewLabel("[no images]")
    }
    // ... other columns
})

table.UpdateCell((col, row, cell) => {
    if (col == 1) {
        // Load the actual image for this row
        let image = state.images[row]
        let absPath = os.getwd() + "/" + image.path
        cell.SetImageFromURI("file://" + absPath)
    }
})
```

## Key Concepts

### Widget Mode Table
- Uses `CreateCell` to create widgets once
- Uses `UpdateCell` to configure widgets with data
- Widgets are cached and reused efficiently

### Canvas Images
- `canvas.NewImageFromFile(path)` - Create image widget
- `cell.SetImageFromURI(uri)` - Load image from file:// URI
- `cell.SetMinSize(width, height)` - Set thumbnail size

### File Operations
- `os.readdir(path)` - List directory contents
- `os.getwd()` - Get current working directory
- String methods: `.lower()`, `.ends_with()`

## Extending the Example

### Add Delete Functionality

```js
cell.OnTapped(() => {
    os.remove(image.path)
    state.images = scanImages()
    table.Refresh()
})
```

### Add File Details

Add a column showing file size or dimensions:
```js
table.Columns(() => ["Filename", "Preview", "Size", "Actions"])
```

### Open with System Viewer

```js
cell.OnTapped(() => {
    // Linux
    os.exec("xdg-open", [image.path])
    // macOS
    // os.exec("open", [image.path])
    // Windows
    // os.exec("start", [image.path])
})
```

## Tips

- Keep images reasonably sized (< 5MB) for good performance
- The `SetMinSize` ensures consistent thumbnail dimensions
- Use the Refresh button after adding new images
- Images are loaded on-demand as cells are displayed

## Troubleshooting

**No images showing:**
- Make sure images are in the `pics/` subdirectory
- Check file extensions are .png, .jpg, .jpeg, or .gif
- Click the Refresh button

**Images too large/small:**
- Adjust `SetMinSize(width, height)` values in UpdateCell
- Adjust column width with `SetColumnWidth()`

**Performance issues:**
- Reduce image file sizes
- Limit the number of images in the directory
- Consider adding pagination (see table.Data offset/limit parameters)
