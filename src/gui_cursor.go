package opennox

import (
	"image"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

const cursorSize = 64

func (c *Client) getCursorAnimFrame(ref *legacy.ImageRef, dt int) *noxrender.Image {
	if ref == nil {
		return nil
	}
	// During a map change the cursor can still point at an image slot that
	// is momentarily not an animation (or has no frames yet); skip drawing
	// this frame instead of panicking in Field24ptr.
	if ref.Kind() != 2 {
		return nil
	}
	anim := ref.Field24ptr()
	ts := int(c.GetInputSeq()) + dt
	imgs := anim.Images()
	if len(imgs) == 0 {
		return nil
	}
	switch anim.AnimType {
	case 0: // OneShot
		ind := (ts - int(anim.Field_3)) / (int(anim.Field_2_1) + 1)
		if ind+1 >= len(imgs) {
			ind = len(imgs) - 1
			if anim.OnEnd != nil {
				ccall.CallVoidPtr(anim.OnEnd, ref.C())
			}
		}
		if ind < 0 { // the input sequence can reset across a map switch
			ind = 0
		}
		return c.r.Bag.AsImage(imgs[ind])
	case 2: // Loop
		ind := ts / (int(anim.Field_2_1) + 1)
		ind %= len(imgs)
		if ind < 0 {
			ind += len(imgs)
		}
		return c.r.Bag.AsImage(imgs[ind])
	default:
		return nil
	}
}

func (c *Client) sub_4BE710(ref *legacy.ImageRef, pos image.Point, ind int) {
	anim := ref.Field24ptr()
	imgs := anim.Images()
	img := c.r.Bag.AsImage(imgs[ind])
	if c.flag3798728 {
		c.r.noxDrawCursor(img, pos)
	} else {
		c.r.DrawImageAt(img, pos)
	}
}
