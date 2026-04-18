import { render, screen } from "@testing-library/react";
import { BookingHoldForm } from "../../src/components/booking/BookingHoldForm";
import { describe, it, expect } from "vitest";

describe("BookingHoldForm", () => {
  it("renders form fields", () => {
    render(
      <BookingHoldForm
        onSubmit={async () => {}}
        packages={[{ id: 1, name: "Package 1" }]}
        fetchSlots={async () => []}
        fetchChairs={async () => []}
        hosts={[{ id: 1, username: "Host 1" }]}
        rooms={[{ id: 1, name: "Room 1" }]}
      />
    );
    expect(screen.getByLabelText(/package/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/host/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/room/i)).toBeInTheDocument();
  });
});
